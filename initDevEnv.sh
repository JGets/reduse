#!/bin/bash

##Simple script to initialize a development environment for Redu.se
export GOPATH="`pwd`/"
export PORT="8080"
export REDUSEDEVELOPMODE="true"

echo "Automated develpment environment setup complete."
echo "Note, you must manually set up the database enviroment variables:"
echo "REDUSE_DB_NAME = $REDUSE_DB_NAME"
echo "REDUSE_DB_ADDRESS = $REDUSE_DB_ADDRESS"
echo "REDUSE_DB_USERNAME = $REDUSE_DB_USERNAME"
echo "REDUSE_DB_PASSWORD = $REDUSE_DB_PASSWORD"
