package rabbitmq

//InitRabbit func
func InitRabbit(uri string, queues Queues) (RabbitImpl, error) {
	rabbit := RabbitImpl{}
	rabbit.Conn = rabbit.GetConnection(uri)
	rabbit.Channel = rabbit.GetChannel()

	err := rabbit.CreateFanoutExchange(ExchangeAddRepo)
	if err != nil {
		return rabbit, err
	}

	if err := createQueues(rabbit, queues); err != nil {
		return rabbit, err
	}

	if err := rabbit.Bind(queues.AddRepoQueue, "", ExchangeAddRepo); err != nil {
		return rabbit, err
	}

	return rabbit, nil
}

func createQueues(rabbit RabbitInterface, queues Queues) error {
	list := []string{queues.DeleteRepoQueue, queues.InstallQueue, queues.ResultInstallQueue, queues.AddRepoQueue}
	for _, name := range list {
		if err := rabbit.CreateQueue(name); err != nil {
			return err
		}
	}
	return nil
}
