package cdp

import "github.com/ecwid/cdp/pkg/devtool"

// Evaluate Evaluates expression on global object.
func (runtime Runtime) evaluate(expression string, contextID int64, async, returnByValue bool) (*devtool.RemoteObject, error) {
	p := &devtool.EvaluatesExpression{
		Expression:    expression,
		ContextID:     contextID,
		AwaitPromise:  !async,
		ReturnByValue: returnByValue,
	}
	result := new(devtool.EvaluatesResult)
	if err := runtime.call("Runtime.evaluate", p, result); err != nil {
		return nil, err
	}
	if result.ExceptionDetails != nil {
		return nil, result.ExceptionDetails
	}
	return result.Result, nil
}

func (runtime Runtime) getProperties(objectID string) ([]*devtool.PropertyDescriptor, error) {
	p := Map{
		"objectId":               objectID,
		"ownProperties":          true,
		"accessorPropertiesOnly": false,
	}
	result := new(devtool.PropertiesResult)
	if err := runtime.call("Runtime.getProperties", p, result); err != nil {
		return nil, err
	}
	if result.ExceptionDetails != nil {
		return nil, result.ExceptionDetails
	}
	return result.Result, nil
}

func (runtime Runtime) callFunctionOn(objectID string, functionDeclaration string, arg ...interface{}) (*devtool.RemoteObject, error) {
	args := make([]devtool.CallArgument, len(arg))
	for i, a := range arg {
		args[i] = devtool.CallArgument{Value: a}
	}
	p := Map{
		"functionDeclaration": functionDeclaration,
		"objectId":            objectID,
		"arguments":           args,
		"awaitPromise":        true,
		"returnByValue":       false,
	}
	result := new(devtool.EvaluatesResult)
	if err := runtime.call("Runtime.callFunctionOn", p, result); err != nil {
		return nil, err
	}
	if result.ExceptionDetails != nil {
		return nil, result.ExceptionDetails
	}
	return result.Result, nil
}

func (runtime Runtime) releaseObject(objectID string) error {
	return runtime.call("Runtime.releaseObject", Map{"objectId": objectID}, nil)
}

// TerminateExecution Terminate current or next JavaScript execution. Will cancel the termination when the outer-most script execution ends
func (runtime Runtime) TerminateExecution() error {
	return runtime.call("Runtime.terminateExecution", nil, nil)
}
