package mail

type SendEmailDto struct {
	Uuid string `json:"uuid" validate:"required"`

	// Email adresses to send the email to
	To []string `json:"to" validate:"dive,email"`

	// Carbon copy: an array of email adresses to send a copy of the email,
	// the only difference between `cc` and `to` is that `to` is the intended original
	// recipients of the email, whereas `cc` is just people that should be notified
	// between the emails, from a technical perspective they work the same
	Cc []string `json:"cc" validate:"dive,email"`

	// Blind carbon copy: similar to the `cc` field but theyre not show to the recipients
	// meaning they cant se the email adresses on `bcc` and be aware of the copies sent
	Bcc []string `json:"bcc" validate:"dive,email"`

	// Reply-To header: most email clients use this to determine the email to reply to
	// when a user opens the email and clicks reply, should be a different email address
	// than the sender, otherwise it would not make a difference
	ReplyToAddresses []string `json:"reply_to_addresses" validate:"dive,email"`

	// Subject header: by default, the text must be 7-bit ASCII due to SMTP limitations,
	// if a different charset is to be used (like UTF-8) specify it in the SubjectCharset
	SubjectText string `json:"subject_text" validate:"required"`

	// Optional email text content: displayed on clients that do not support Html
	BodyText string `json:"body_text"`

	// Email html content
	BodyHtml string `json:"body_html" validate:"required"`
}

type SendEmailRes struct {
	Success bool `json:"success"`
	// Generic message describring the success or error
	Message string `json:"message"`
}
