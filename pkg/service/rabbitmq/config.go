package rabbitmq

//InitRabbit func
func InitRabbit(uri string, queues Queues) (RabbitImpl, error) {
	rabbit := RabbitImpl{Queues: queues}
	rabbit.Conn = rabbit.GetConnection(uri)
	rabbit.Channel = rabbit.GetChannel()

	err := rabbit.CreateFanoutExchange(ExchangeAddRepo)
	if err != nil {
		return rabbit, err
	}

	err = rabbit.CreateFanoutExchange(ExchangeDelRepo)
	if err != nil {
		return rabbit, err
	}

	err = rabbit.CreateFanoutExchange(ExchangeUpdateRepo)
	if err != nil {
		return rabbit, err
	}

	if err := createQueues(rabbit, queues); err != nil {
		return rabbit, err
	}

	if err := rabbit.Bind(queues.AddRepoQueue, "", ExchangeAddRepo); err != nil {
		return rabbit, err
	}

	if err := rabbit.Bind(queues.DeleteRepoQueue, "", ExchangeDelRepo); err != nil {
		return rabbit, err
	}

	if err := rabbit.Bind(queues.UpdateRepoQueue, "", ExchangeUpdateRepo); err != nil {
		return rabbit, err
	}

	return rabbit, nil
}

func createQueues(rabbit RabbitInterface, queues Queues) error {
	listExclusive := []string{queues.DeleteRepoQueue, queues.AddRepoQueue, queues.UpdateRepoQueue}
	for _, name := range listExclusive {
		if err := rabbit.CreateQueue(name, true); err != nil {
			return err
		}
	}
	listNoExclusive := []string{queues.InstallQueue, queues.ResultInstallQueue}
	for _, name := range listNoExclusive {
		if err := rabbit.CreateQueue(name, false); err != nil {
			return err
		}
	}
	return nil
}
