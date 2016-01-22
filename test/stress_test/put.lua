require('io')
require('sha1')
require('base64')


oids={}
for j = 1,1000,1
do
	oids[j] = "oid"..j

end

i=0

file = io.open('/home/zhangdongmao/CLOUD-2013-20130619-1371604329974.mp4', "rb")
if file == nil then 
	print("error")
	return nil 
end
content = file:read "*a"
file:close()


--
--request = function(){}
--
function request()
	local x = i
	x = x % #oids + 1
	i = i + 1
	local path = "/video/"..oids[x]
	local key = to_base64(hmac_sha1_binary("wuzei",path))
	local header = {["Authorization"] = key}
	return wrk.format("PUT", path, header, content)
end
