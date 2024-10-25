- **zoha-lmtp** An example LMTP client that should be installed on each node that stores EML files
- **zoha-sender-client**  Forwards messages created by the service "send copy to" commands, autoresponder
 or aliases, to the main zoha-sender-server
- **zoha-sender-server**  A master Sender server, forwarding messages from Zoha-Sender-Client to the main MTA (e.g. the main instance of postfix), here the whole fun starts again and again the MTA can forward the message to ZOHA-LMTP
- **zoha-route** a proxy client for nginx that forwards incoming traffic to the appropriate IMAP, POP3, SMTP instances
