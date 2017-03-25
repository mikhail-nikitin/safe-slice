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
	action int
	value interface{}
	index int
	result chan interface{}
	updater UpdateFunc
}

type searchResult struct {
	value interface{}
	found bool
}

func (s safeSlice) Append(value interface{}) {
	s <- command{action: appendElement, value: value}
}

func (s safeSlice) At(index int) interface{} {
	resultChannel := make(chan interface{})
	s <- command{action: getElementAt, index: index, result: resultChannel}
	sr := (<-resultChannel).(searchResult)
	if sr.found {
		return sr.value
	} else {
		return nil
	}
}

func (s safeSlice) Close() []interface{} {
	resultChannel := make(chan interface{})
	s <- command{action: end, result: resultChannel }
	return (<- resultChannel).([]interface{})
}

func (s safeSlice) Delete(index int) {
	s <- command{action: deleteElementAt, index: index }
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
			cmd := <- s
			switch (cmd.action) {
			case appendElement:
				data = append(data, cmd.value)
			case getElementAt:
				r := searchResult{}
				if len(data) > cmd.index && cmd.index >= 0 {
					r.value = data[cmd.index]
					r.found = true
				}
				cmd.result <- r
			case deleteElementAt:
				if len(data) > cmd.index && cmd.index >= 0 {
					data = append(data[:cmd.index], data[cmd.index + 1:]...)
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