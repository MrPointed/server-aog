package actions

type Action[S any] func(service S)

type ActionExecutor[S any] struct {
	service S
	actions chan Action[S]
}

func NewActionExecutor[S any](service S) *ActionExecutor[S] {
	return &ActionExecutor[S]{
		service: service,
		actions: make(chan Action[S], 100),
	}
}

func (e *ActionExecutor[S]) Start() {
	go func() {
		for action := range e.actions {
			action(e.service)
		}
	}()
}

func (e *ActionExecutor[S]) Dispatch(action Action[S]) {
	e.actions <- action
}