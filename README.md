
# Instant Pillar

This is a Telegram bot sending random photos to a user. Photos are being taken from Pinterest boards.

## Requirements

- Go language

- PostgreSQL

### Dependencies

-  [go-telegram-bot-api](https://github.com/go-telegram-bot-api) @github/Syfaro Go client for Telegram Bot API

-  [go-pinterest](https://github.com/a-frony/go-pinterest) working fork of @github/carrot Go client for Pinterest API

## Settings

Use `config.json` to fill settings:

-  `Language`: preffered language. English and Russian languages included

-  `TelegramBotToken`: get this token from [@BotFather](https://t.me/BotFather_bot) bot in Telegram

-  `PinterestToken`: get this token from Pinterest application. Check out their [getting started](https://developers.pinterest.com/docs/api/overview/) guide. Note that your application cannot be submitted, so you will have limitations. You can upload to the bot up to 500 photos per hour 

-  `BotAdmin`: your Telegram username. This user can upload new photos to the bot

-  `DBUser, DBPass, DBName`: connect parameters to PostgreSQL

## Working with bot

Administrator can use `/load` command. The bot will wait for a correct Pinterest board link like https://pinterest.com/user/board/ and fetch photos from this board.

Users can use commands `/start` or `/moar` to get new random photo.

Check out working instance [@instantPillars](https://t.me/instantPillars_bot) with photos of cute caterpillars.

  

## License

  

[MIT](LICENSE.md) © 2019 a-frony

