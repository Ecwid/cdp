package cdp

// TargetInfo https://chromedevtools.github.io/devtools-protocol/tot/Target#type-TargetInfo
type TargetInfo struct {
	TargetID         string `json:"targetId"`
	Type             string `json:"type"`
	Title            string `json:"title"`
	URL              string `json:"url"`
	Attached         bool   `json:"attached"`
	OpenerID         string `json:"openerId"`
	BrowserContextID string `json:"browserContextId"`
}

type targetCrashed struct {
	TargetID  string `json:"targetId"`
	Status    string `json:"status"`
	ErrorCode int64  `json:"errorCode"`
}

func (session *Session) getTargets() ([]TargetInfo, error) {
	msg, err := session.blockingSend("Target.getTargets", &Params{})
	if err != nil {
		return nil, err
	}
	var targets []TargetInfo
	unmarshal(msg["targetInfos"], &targets)
	return targets, nil
}

func (session *Session) attachToTarget(targetID string) (string, error) {
	msg, err := session.blockingSend("Target.attachToTarget", &Params{"targetId": targetID, "flatten": true})
	if err != nil {
		return "", err
	}
	return msg["sessionId"].(string), nil
}

func (session *Session) detachFromTarget(sessionID string) error {
	_, err := session.blockingSend("Target.detachFromTarget", &Params{"sessionId": sessionID})
	return err
}

func (session *Session) setAutoAttach(autoAttach bool) error {
	_, err := session.blockingSend("Target.setAutoAttach", &Params{
		"autoAttach":             autoAttach,
		"waitForDebuggerOnStart": false,
		"flatten":                true,
	})
	return err
}

func (session *Session) setDiscoverTargets(discover bool) error {
	_, err := session.blockingSend("Target.setDiscoverTargets", &Params{"discover": true})
	return err
}

func (session *Session) activateTarget(targetID string) error {
	_, err := session.blockingSend("Target.activateTarget", &Params{"targetId": targetID})
	return err
}

func (session *Session) createTarget(url string) (string, error) {
	msg, err := session.blockingSend("Target.createTarget", &Params{
		"url": url,
	})
	if err != nil {
		return "", err
	}
	return msg["targetId"].(string), nil
}

func (session *Session) closeTarget(targetID string) (bool, error) {
	msg, err := session.blockingSend("Target.closeTarget", &Params{"targetId": targetID})
	// event 'Target.targetDestroyed' was received early than message response
	if err == ErrSessionClosed {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return msg["success"].(bool), nil
}
