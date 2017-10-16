
const sys = require('util');
const exec = require('child_process').exec;

const config = require('./config/env.json');
const SLACK_CHANNEL = config.SLACK_CHANNEL;
const SLACK_TOKEN = config.SLACK_TOKEN;

const app_cmd = `SLACK_CHANNEL=${SLACK_CHANNEL} SLACK_TOKEN=${SLACK_TOKEN} CONFIG_PATH=config/checks.json go run ./cmd/main.go`;

let app_start = exec(app_cmd);
