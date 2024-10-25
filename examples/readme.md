The folders contain sample configurations that will allow you to run full mail serwer (smtp/pop3/imap4) with Zoha (Golang) as an LMTP server

 - **postfix/** configuring a Postfix server on the main machine responsible for receiving and sending mail through your entire system
 - **courier/** courier configuration files IMAP4/POP3 listening for connections from route66
 - **nginx/** nginx proxy imap/pop3 to route66