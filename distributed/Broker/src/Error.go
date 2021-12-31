package BrokerService

type ChannelExist struct {
}

func (e *ChannelExist) Error() (s string) {
	return "Id already Exist"
}

type IdDoesNotExist struct {
}

func (e *IdDoesNotExist) Error() (s string) {
	return "ID Does Not Exist"
}
