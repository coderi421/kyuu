package session

import (
	"github.com/coderi421/kyuu"
)

// Manager 为了简化使用，提供了一些常用的方法
type Manager struct {
	Store
	Propagator
	SessCtxKey string // 在 context 中的备份，方便使用
}

// GetSession attempts to retrieve the Session from the context.
// If successful, it caches the Session instance in the UserValues of the context.
func (m *Manager) GetSession(ctx *kyuu.Context) (Session, error) {
	// Create UserValues map if it doesn't exist
	// Transfer data between middlewares
	if ctx.UserValues == nil {
		ctx.UserValues = make(map[string]interface{}, 1)
	}

	// Try to retrieve Session from cache
	val, ok := ctx.UserValues[m.SessCtxKey]
	if ok {
		return val.(Session), nil
	}

	// Extract session ID from the request
	id, err := m.Extract(ctx.Req)
	if err != nil {
		return nil, err
	}

	// Fetch session from storage
	sess, err := m.Get(ctx.Req.Context(), id)
	if err != nil {
		return nil, err
	}

	// Cache session in UserValues
	ctx.UserValues[m.SessCtxKey] = sess
	return sess, nil
}

// InitSession initializes a session with the given id.
// It generates a new session and injects the id into the context response.
func (m *Manager) InitSession(ctx *kyuu.Context, id string) (Session, error) {
	sess, err := m.Generate(ctx.Req.Context(), id)
	if err != nil {
		return nil, err
	}

	// injects the id into the context response
	if err = m.Inject(id, ctx.Resp); err != nil {
		return nil, err
	}

	return sess, nil
}

// RefreshSession refreshes the session using the provided context.
// It returns the updated session or an error if the refresh fails.
func (m *Manager) RefreshSession(ctx *kyuu.Context) (Session, error) {
	// Get the session
	sess, err := m.GetSession(ctx)
	if err != nil {
		return nil, err
	}

	// Refresh the session
	if err = m.Refresh(ctx.Req.Context(), sess.ID()); err != nil {
		return nil, err
	}

	// Inject the updated session
	if err = m.Inject(sess.ID(), ctx.Resp); err != nil {
		return nil, err
	}

	return sess, nil
}

// RemoveSession removes the session associated with the given context.
// It returns an error if the session cannot be found or removed.
func (m *Manager) RemoveSession(ctx *kyuu.Context) error {
	// Get the session from the context
	sess, err := m.GetSession(ctx)
	if err != nil {
		return err
	}

	// Remove the session from the server store
	if err = m.Store.Remove(ctx.Req.Context(), sess.ID()); err != nil {
		return err
	}

	// Remove the session ID from the cookie or header...
	return m.Propagator.Remove(ctx.Resp)
}
