package cdp

func (session Session) Animation(enable bool) error {
	if enable {
		return session.call("Animation.enable", nil, nil)
	} else {
		return session.call("Animation.disable", nil, nil)
	}
}
