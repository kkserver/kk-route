
function OnMessage( message, from )
	print(message.To)
	print(kk)
	print(kk.NewMessage)
	print(kk.NewMessage("DONE"))
	return message.To == "kk.job.job/slave/process"
end
