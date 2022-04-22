## A WhatsApp to Signal bridge
[Docker image](https://hub.docker.com/repository/docker/oskbor/whatsapp-signal-bridge/general)
**Don't run the image in the cloud, you will get your whatsapp account banned and we both will be very sad.**
Is supports text messages, attachments, images, audio and video. Sending attachments from Signal to WhatsApp has not been verified since I use a Punkt MP02 as my signal device.

### How to run
 1. [Setup the signal container](https://github.com/bbernhard/signal-cli-rest-api#getting-started)
 2. Start the signal container in `json-rpc` mode.
 3. Start the bridge container. 
 4. On first startup a QR code will be seen in `stdout`. Scan it with whatsapp on your phone to login.
 5. If all is well, it should now be working!

### Todo

- [ ] Simplify the setup procedure
- [ ] Verify sending attachements from signal to whatsapp
- [ ] Notify on group changes (members added/removed etc)
- [ ] Typing indicators? Read reciepts? Etc
