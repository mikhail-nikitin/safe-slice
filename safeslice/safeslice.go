package safeslice

type SafeSlice interface {
	Append(interface{})
	At(int) interface{}
	Close() []interface{}
	Delete(int)
	Len() int
	Update(int, UpdateFunc)
}

type UpdateFunc func(interface{}) interface{}

type safeSlice chan command

const (
	appendElement   = iota
	getElementAt
	deleteElementAt
	getLength
	update
	end
)

type command struct {
	action  int
	value   interface{}
	index   int
	result  chan interface{}
	updater UpdateFunc
}

func (s safeSlice) Append(value interface{}) {
	s <- command{action: appendElement, value: value}
}

func (s safeSlice) At(index int) interface{} {
	resultChannel := make(chan interface{})
	s <- command{action: getElementAt, index: index, result: resultChannel}
	return <-resultChannel
}

func (s safeSlice) Close() []interface{} {
	resultChannel := make(chan interface{})
	s <- command{action: end, result: resultChannel}
	return (<-resultChannel).([]interface{})
}

func (s safeSlice) Delete(index int) {
	s <- command{action: deleteElementAt, index: index}
}

func (s safeSlice) Len() int {
	resultChannel := make(chan interface{})
	s <- command{action: getLength, result: resultChannel}
	return (<-resultChannel).(int)
}

func (s safeSlice) Update(index int, updater UpdateFunc) {
	s <- command{action: update, updater: updater, index: index}
}

func New() SafeSlice {
	s := make(safeSlice)
	s.run()
	return s
}

func (s safeSlice) run() {
	go func() {
		data := make([]interface{}, 0)
		for {
			cmd := <-s
			switch cmd.action {
			case appendElement:
				data = append(data, cmd.value)
			case getElementAt:
				var result interface{} = nil
				if len(data) > cmd.index && cmd.index >= 0 {
					result = data[cmd.index]
				}
				cmd.result <- result
			case deleteElementAt:
				if len(data) > cmd.index && cmd.index >= 0 {
					head := data[:cmd.index]
					tail := data[cmd.index+1:]
					data = append(head, tail...)
				}
			case getLength:
				cmd.result <- len(data)
			case update:
				if len(data) > cmd.index && cmd.index >= 0 {
					data[cmd.index] = cmd.updater(data[cmd.index])
				}
			case end:
				cmd.result <- data
				close(s)
				return
			}
		}
	}()
}
