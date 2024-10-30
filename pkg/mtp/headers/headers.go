package headers

const (
	/*

		The “envelope from” is a return path, which is the return address hidden in the email message header that instructs
		mail servers or inbox service providers (ISPs) where to return messages if they bounce. This address is used for email delivery.
	*/
	XEnvelopeFrom = "X-Envelope-From"
	/*
		Postfix:
		  In  order  to  stop  mail  forwarding loops early, the software adds an
		       optional  Delivered-To:  header  with  the  final  envelope   recipient
		       address.  If  mail  arrives for a recipient that is already listed in a
		       Delivered-To: header, the message is bounced.
	*/
	DeliveredTo = "Delivered-To"

	// XStoredAt On which server we saved the file (hostname)
	XStoredAt = "X-Stored-At"
	// XServerosTag Tags to recognize the message type
	// [TAG][TAG]
	XServerosTag = "X-Serveros-Tag"
	XAction      = "X-Action"

	// XLoop fixme we should replace this with DeliveredTo - because that header seems to meet this assumption
	XLoop = "X-Loop"
)
