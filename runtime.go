package cdp

import (
	"github.com/ecwid/cdp/pkg/devtool"
)

// Evaluate evaluate javascript code at context of web page
func (session Runtime) Evaluate(code string, async bool, returnByValue bool) (interface{}, error) {
	result, err := session.evaluate(code, session.currentContext(), async, returnByValue)
	if err != nil {
		return "", err
	}
	return result.Value, nil
}

// Evaluate Evaluates expression on global object.
func (session Runtime) evaluate(expression string, contextID int64, async, returnByValue bool) (*devtool.RemoteObject, error) {
	p := &devtool.EvaluatesExpression{
		Expression:            expression,
		IncludeCommandLineAPI: true,
		ContextID:             contextID,
		AwaitPromise:          !async,
		ReturnByValue:         returnByValue,
	}
	result := new(devtool.EvaluatesResult)
	if err := session.call("Runtime.evaluate", p, result); err != nil {
		return nil, err
	}
	if result.ExceptionDetails != nil {
		return nil, result.ExceptionDetails
	}
	return result.Result, nil
}

func (session Runtime) getProperties(objectID string) ([]*devtool.PropertyDescriptor, error) {
	p := Map{
		"objectId":               objectID,
		"ownProperties":          true,
		"accessorPropertiesOnly": false,
	}
	result := new(devtool.PropertiesResult)
	if err := session.call("Runtime.getProperties", p, result); err != nil {
		return nil, err
	}
	if result.ExceptionDetails != nil {
		return nil, result.ExceptionDetails
	}
	return result.Result, nil
}

func (session Runtime) callFunctionOn(objectID string, functionDeclaration string, awaitPromise, returnByValue bool, arg ...interface{}) (*devtool.RemoteObject, error) {
	args := make([]devtool.CallArgument, len(arg))
	for i, a := range arg {
		args[i] = devtool.CallArgument{Value: a}
	}
	p := Map{
		"functionDeclaration": functionDeclaration,
		"objectId":            objectID,
		"arguments":           args,
		"awaitPromise":        awaitPromise,
		"returnByValue":       returnByValue,
	}
	result := new(devtool.EvaluatesResult)
	if err := session.call("Runtime.callFunctionOn", p, result); err != nil {
		return nil, err
	}
	if result.ExceptionDetails != nil {
		return nil, result.ExceptionDetails
	}
	return result.Result, nil
}

func (session Runtime) releaseObject(objectID string) error {
	return session.call("Runtime.releaseObject", Map{"objectId": objectID}, nil)
}

// TerminateExecution Terminate current or next JavaScript execution. Will cancel the termination when the outer-most script execution ends
func (session Runtime) TerminateExecution() error {
	return session.call("Runtime.terminateExecution", nil, nil)
}
