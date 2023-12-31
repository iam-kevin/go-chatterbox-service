# chatterbox-go

Service that's powering [chatterbox](https://chatterbox-kevin.vercel.app/).

## About

For the purpose of building confidence in the Golang language and Cloud Native environment, this service was created to help achive such feat.

This service helps powering all the server activities that's to take place in the chatterbox. From fetching user information to sending messages via socket connection

## Deployment

This service is currently being deployed to my personal Fly.io App instance. [Live link of service](https://chatterbox-go.fly.dev/)

CD pipeline is set such that, any changes made to the code, would rebuild the image serve the code to the same link
