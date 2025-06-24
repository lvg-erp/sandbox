package waitg

type WaitG interface {
	Add(int)
	Done()
	Wait()
}
