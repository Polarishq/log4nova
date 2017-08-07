#Log4Nova
A simple logging library for forwarding your logs into your Nova log store

##Setting up Log4Nova
1. Vendor this library into your go-service
2. Wire up the NovaHandler to your middleware chain with your Nova API Keys:
```
//Example for standing up the log4nova handler
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	clientID := os.Getenv("NOVA_CLIENT_ID")
	clientSecret := os.Getenv("NOVA_CLIENT_SECRET")
	novaLogger := log4nova.NewNovaLogger(clientID, clientSecret)
    log4nova.NewNovaHandler(novaLogger,
        handlers.SomeOtherHandler(handler))
}
``` 

##Usage
To log any data to your splunk instance, first initialize a log4nova instance:
```
	clientID := os.Getenv("NOVA_CLIENT_ID")
	clientSecret := os.Getenv("NOVA_CLIENT_SECRET")
	novaLogger := log4nova.NewNovaLogger(clientID, clientSecret)
	
	//Batch logs every 5 secs
	novaLogger.SendInterval = 5000
	
	//Mirror all data to stdout
	novaLogger.SendToStdOut = true
	
	//Start the logger.  If the logger is not started, logs will not be sent to the logstore
	novaLogger.Start()
	
	//Send my logs off to nova
	novaLogger.WithFields({
	    SomeField: "Some important value"    
    }).Errorf("Error! %v", e)
    
    //Stop the logger
    novaLogger.Stop()

```

##Troubleshooting
TBD