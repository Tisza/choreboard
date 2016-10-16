var Botkit = require('botkit');
const fs = require('fs');
var slack_api_token = fs.readFileSync('/Users/johndowling/GoWorkspace/src/choreboard/slackbot/slack_token.txt').toString();
// console.log(slack_api_token.slice(0, slack_api_token.length-1));

var controller = Botkit.slackbot({
  debug: false
  //include "log: false" to disable logging
  //or a "logLevel" integer from 0 to 7 to adjust logging verbosity
});

// connect the bot to a stream of messages
controller.spawn({
  token: slack_api_token.slice(0, slack_api_token.length-1),
  incoming_webhook: {
    url: "https://hooks.slack.com/services/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"
  }
}).startRTM()

// give the bot something to listen for.
controller.hears('hello',['direct_message','direct_mention','mention'],function(bot,message) {
  bot.reply(message,'Hello yourself.');
});

controller.hears('show me the (.*)',['message_received', 'direct_message','direct_mention','mention'],function(bot,message) {
  var thing = message.match[1]; //match[1] is the (.*) group. match[0] is the entire group (open the (.*) doors).
  if (thing === 'scores') {
    return bot.reply(message, "Mkay, here's everyone's scores...");
  } else if (thing == 'chores') {
    return bot.reply(message, "Mkay, here's the status of all the chores...");
  } else if (thing == 'webhook') {
    bot.sendWebhook({
      text: 'This is an incoming webhook',
    });
    return;
  } else if (thing == 'butter') {
    return bot.reply(message, "MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM");
  }
  return bot.reply(message, "NO! I don't want that...");
});

//Using attachments
controller.hears('attachments',['message_received', 'direct_message','direct_mention','mention'],function(bot,message) {
  var reply_with_attachments = {
    'username': 'My bot' ,
    'text': 'This is a pre-text',
    'attachments': [
      {
        'fallback': 'To be useful, I need you to invite me in a channel.',
        'title': 'How can I help you?',
        'text': 'To be useful, I need you to invite me in a channel ',
        'color': '#7CD197'
      }
    ],
    'icon_url': 'http://lorempixel.com/48/48'
    }

  bot.reply(message, reply_with_attachments);
});
