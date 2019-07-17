package cdp

import (
	"fmt"
)

// ExecutionContextDescription https://chromedevtools.github.io/devtools-protocol/tot/Runtime#type-ExecutionContextDescription
type ExecutionContextDescription struct {
	ID      int64                  `json:"id"`
	Origin  string                 `json:"origin"`
	Name    string                 `json:"name"`
	AuxData map[string]interface{} `json:"auxData"`
}

// RemoteObject https://chromedevtools.github.io/devtools-protocol/tot/Runtime#type-RemoteObject
type RemoteObject struct {
	Type                string      `json:"type"`
	Subtype             string      `json:"subtype"`
	ClassName           string      `json:"className"`
	Value               interface{} `json:"value"`
	UnserializableValue string      `json:"unserializableValue"`
	Description         string      `json:"description"`
	ObjectID            string      `json:"objectId"`
}

// ExceptionDetails https://chromedevtools.github.io/devtools-protocol/tot/Runtime#type-ExceptionDetails
type ExceptionDetails struct {
	ExceptionID        int64         `json:"exceptionId"`
	Text               string        `json:"text"`
	LineNumber         int64         `json:"lineNumber"`
	ColumnNumber       int64         `json:"columnNumber"`
	ScriptID           string        `json:"scriptId"`
	URL                string        `json:"url"`
	Exception          *RemoteObject `json:"exception"`
	ExecutionContextID int64         `json:"executionContextId"`
}

// PropertyDescriptor https://chromedevtools.github.io/devtools-protocol/tot/Runtime#type-PropertyDescriptor
type PropertyDescriptor struct {
	Name         string        `json:"name"`
	Value        RemoteObject  `json:"value"`
	Writable     bool          `json:"writable"`
	Get          *RemoteObject `json:"get"`
	Set          *RemoteObject `json:"set"`
	Configurable bool          `json:"configurable"`
	Enumerable   bool          `json:"enumerable"`
	WasThrown    bool          `json:"wasThrown"`
	IsOwn        bool          `json:"isOwn"`
	Symbol       *RemoteObject `json:"symbol"`
}

// CallArgument https://chromedevtools.github.io/devtools-protocol/tot/Runtime#type-CallArgument
type CallArgument struct {
	Value               interface{} `json:"value,omitempty"`
	UnserializableValue string      `json:"unserializableValue,omitempty"`
	ObjectID            string      `json:"objectId,omitempty"`
}

// EvaluatesExpression https://chromedevtools.github.io/devtools-protocol/tot/Runtime#method-evaluate
type EvaluatesExpression struct {
	Expression            string `json:"expression"`
	ObjectGroup           string `json:"objectGroup,omitempty"`
	IncludeCommandLineAPI bool   `json:"includeCommandLineAPI,omitempty"`
	Silent                bool   `json:"silent,omitempty"`
	ContextID             int64  `json:"contextId,omitempty"`
	ReturnByValue         bool   `json:"returnByValue,omitempty"`
	GeneratePreview       bool   `json:"generatePreview,omitempty"`
	UserGesture           bool   `json:"userGesture,omitempty"`
	AwaitPromise          bool   `json:"awaitPromise,omitempty"`
	ThrowOnSideEffect     bool   `json:"throwOnSideEffect,omitempty"`
	Timeout               int64  `json:"timeout,omitempty"`
}

func (r *RemoteObject) bool() bool {
	return r.Type == "boolean" && r.Value.(bool)
}

func (e *ExceptionDetails) Error() string {
	return fmt.Sprintf("%+v", e.Exception)
}

func (session *Session) getExceptionDetails(msg MessageResult) error {
	if exceptionDetails, has := msg["exceptionDetails"]; has {
		details := &ExceptionDetails{}
		unmarshal(exceptionDetails, details)
		return details
	}
	return nil
}

// https://chromedevtools.github.io/devtools-protocol/tot/Runtime#method-releaseObject
func (session *Session) releaseObject(objectID string) error {
	_, err := session.blockingSend("Runtime.releaseObject", &Params{"objectId": objectID})
	return err
}

// Evaluate Evaluates expression on global object.
func (session *Session) Evaluate(expression string, contextID int64) (*RemoteObject, error) {
	exp := &EvaluatesExpression{
		Expression:    expression,
		ContextID:     contextID,
		AwaitPromise:  true,
		ReturnByValue: false,
	}
	params := &Params{}
	unmarshal(exp, params)
	res, err := session.blockingSend("Runtime.evaluate", params)
	if err != nil {
		return nil, err
	}
	if err := session.getExceptionDetails(res); err != nil {
		return nil, err
	}
	obj := &RemoteObject{}
	unmarshal(res["result"], obj)
	return obj, nil
}

func (session *Session) getProperties(objectID string) ([]*PropertyDescriptor, error) {
	evaluated, err := session.blockingSend("Runtime.getProperties", &Params{
		"objectId":               objectID,
		"ownProperties":          true,
		"accessorPropertiesOnly": false,
	})
	if err != nil {
		return nil, err
	}
	err = session.getExceptionDetails(evaluated)
	if err != nil {
		return nil, err
	}
	obj := make([]*PropertyDescriptor, 0)
	unmarshal(evaluated["result"], &obj)
	return obj, nil
}

func (session *Session) callFunctionOn(objectID string, functionDeclaration string, arg ...interface{}) (*RemoteObject, error) {
	args := make([]CallArgument, len(arg))
	for i, a := range arg {
		args[i] = CallArgument{Value: a}
	}
	evaluated, err := session.blockingSend("Runtime.callFunctionOn", &Params{
		"functionDeclaration": functionDeclaration,
		"objectId":            objectID,
		"arguments":           args,
		"awaitPromise":        true,
		"returnByValue":       false,
	})
	if err != nil {
		return nil, err
	}
	err = session.getExceptionDetails(evaluated)
	if err != nil {
		return nil, err
	}
	obj := &RemoteObject{}
	unmarshal(evaluated["result"], obj)
	return obj, nil
}
