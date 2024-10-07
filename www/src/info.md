
## More information 

{{ .Config.MailFullyQualifiedDomainName }} provides a low-maintenance, resource efficient and 
interoperable e-mail service for everyone. What's behind a `chatmail` is 
effectively a normal e-mail address just like any other but optimized 
for usage in chats, especially DeltaChat.

### Choosing a chatmail address instead of using a random one

By default, signing up using the QR code or invite link will generate a random address. For most use cases, this is fine---Delta Chat prefers showing the name you configure for yourself, rather than the email address.  However, if you would like to choose your own email address, it is possible to configure.  In the Delta Chat account setup you may tap `Create a profile` then `Use other server` and choose `Classic e-mail login`. Here, fill the two fields like this: 

- `E-Mail Address`: invent a word with
{{ if eq .Config.UsernameMinLength .Config.UsernameMaxLength }}
  *exactly* {{ .Config.UsernameMinLength }}
{{ else }}
  {{ .Config.UsernameMinLength }}
  {{ if gt .Config.UsernameMaxLength 12 }}
    or more
  {{ else }}
    to {{ .Config.UsernameMaxLength }}
  {{ end }}
{{ end }}
  characters
  and append `@{{.Config.MailFullyQualifiedDomainName}}` to it.

- `Existing Password`: invent at least {{ .Config.PasswordMinLength }} characters.

If the e-mail address is not yet taken, you'll get that account. 
The first login sets your password. 


### Rate and storage limits 

- Un-encrypted messages are blocked to recipients outside
  {{.Config.MailFullyQualifiedDomainName}} but setting up contact via [QR invite codes](https://delta.chat/en/help#howtoe2ee) 
  allows your messages to pass freely to any outside recipients.

- You may send up to {{ .Config.MaxEmailsPerMinutePerUser }} messages per minute.

- Messages are unconditionally removed {{ .Config.DeleteMailsAfterDays }} days after arriving on the server.

- You can store up to [{{ .Config.MaxMailboxSizeMB }} megabytes of messages on the server](https://delta.chat/en/help#what-happens-if-i-turn-on-delete-old-messages-from-server).


### <a name="account-deletion"></a> Account deletion 

If you remove a {{ .Config.MailFullyQualifiedDomainName }} profile from within the Delta Chat app, 
then the account on the server, along with all associated data, is automatically
deleted {{ .Config.DeleteInactiveUsersAfterDays }} days afterwards. 

If you use multiple devices, the account on the server will be deleted {{
.Config.DeleteInactiveUsersAfterDays }} days after you remove the chat profile
from *every* device.

If you have any further questions or requests regarding account deletion
please send a message from your account to {{ .Config.PrivacyContactEmailAddress }}. 


### Who are the operators? Which software is running? 

This chatmail provider is run by a small voluntary group of devs and sysadmins,
who [publically develop chatmail provider setups](https://github.com/deltachat/chatmail).
Chatmail setups aim to be very low-maintenance, resource efficient and 
interoperable with any other standards-compliant e-mail service. 
