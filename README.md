## Golang GroupMe Bot

This project serves as the basis for running an HTTP server that listens for messages on a [GroupMe](https://web.groupme.com/) group and handling running commands and responding to the messages if given trigger symbols or terms are found within the message.

For more information about GroupMe bots, the [GroupMe bots tutorial](https://dev.groupme.com/tutorials/bots) is a helpful resource.

## Installation

`go get github.com/sha1sum/golang_groupme_bot`

## Basic Usage

 1. [Create a GroupMe bot](https://dev.groupme.com/bots)
 2. Set environment variable `GROUPME_BOT_ID` to the ID of the bot that you created in step 1
 3. If using Heroku, the HTTP server is already set to use the port defined in the `PORT` environment variable. If you're not using Heroku, you can set the port by assigning the number to the environment variable, or stick with the default of `80`.
 4. Create a new Go application and your handler. Your handler should implement the `bot.Handler` interface by defining one function: `Handle(term string, c chan []*OutgoingMessage, message IncomingMessage)`. This function will allow you access to the full `IncomingMessage` as parsed by the GroupMe bot callback post. `term` is the text of the message that was posted with instances of your `Command.Trigger` removed from the text (as well as surrounding spaces).
 5. In your application, create at least one `Command` and append it to a slice (`[]bot.Command`). Define a slice of strings that, when found, will fire off the `Handle` function in your `bot.Handler`. Set this slice as the value for `Command.Triggers`. Make a new instance of your `Handler` and set that instance as the value for `Command.Handler`.
 6. Construct an `OutgoingMessage` (see `OutgoingMessage` in `bot/bot.go`) and send it to the channel given to the `Handle` function in your `bot.Handler` 
 
## Sample Bots

While this project handles the parsing and running of commands, no actual commands are built into the project. For example bot handlers, See the [sha1sum/distinguished_taste_society_bots](https://github.com/sha1sum/distinguished_taste_society_bots) repository.

## Tests

Built-in tests are present for ensuring the HTTP handler returns a 200 ("OK") status code and that incoming bot callbacks can be successfully parsed into messages. To run the tests, simply run:

`go test github.com/sha1sum/golang_groupme_bot/bot`

## License

This project is completely open for forking/copying/modifying. I would appreciate attribution if possible when you use this code.