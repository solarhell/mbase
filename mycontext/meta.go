package mycontext

type keySession struct{}

// SetSession use to set real session name
func SetSession(ctx MyContext, session string) {
	ctx.Env().Set(keySession{}, session)
}

// GetSession use to get session name if it has a real session, otherwise return myContextImpl name instead
func GetSession(ctx MyContext) string {
	if val, ok := ctx.Env().GetString(keySession{}); ok {
		if val != "" {
			return val
		}
	}
	return ctx.Name()
}

// GetRealSession use to get real session name
func GetRealSession(ctx MyContext) (string, bool) {
	return ctx.Env().GetString(keySession{})
}
