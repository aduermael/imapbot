-- Dockerscript
-- A Dockerscript is a script written in Lua, executed in the Docker sandbox.
-- Documentation: http://dockerproj.duermael.com

-- Lists Docker entities involved in project
function status()
	local dockerhost = os.getEnv("DOCKER_HOST")
	if dockerhost == "" then
		dockerhost = "local"
	end
	print("Docker host: " .. dockerhost)

	local success, services = pcall(docker.service.list, '--filter label=docker.project.id:' .. docker.project.id)
	local swarmMode = success

	if swarmMode then
		print("Services:")
		if #services == 0 then
			print("none")
		else
			for i, service in ipairs(services) do
				print(" - " .. service.name .. " image: " .. service.image)
			end
		end
	else
		local containers = docker.container.list('-a --filter label=docker.project.id:' .. docker.project.id)
		print("Containers:")
		if #containers == 0 then
			print("none")
		else
			for i, container in ipairs(containers) do
				print(" - " .. container.name .. " (" .. container.status .. ") image: " .. container.image)
			end
		end
	end

	local volumes = docker.volume.list('--filter label=docker.project.id:' .. docker.project.id)
	print("Volumes:")
	if #volumes == 0 then
		print("none")
	else
		for i, volume in ipairs(volumes) do
			print(" - " .. volume.name .. " (" .. volume.driver .. ")")
		end
	end

	local networks = docker.network.list('--filter label=docker.project.id:' .. docker.project.id)
	print("Networks:")
	if #networks == 0 then
		print("none")
	else
		for i, network in ipairs(networks) do
			print(" - " .. network.name .. " (" .. network.driver .. ")")
		end
	end

	local images = docker.network.list('--filter label=docker.project.id:' .. docker.project.id)
	print("Images (built within project):")
	if #networks == 0 then
		print("none")
	else
		for i, image in ipairs(images) do
			print(" - " .. image.name)
		end
	end
end

----------------
-- UTILS
----------------

utils = {}

-- returns a string combining strings from  string array in parameter
-- an optional string separator can be provided.
utils.join = function(arr, sep)
	str = ""
	if sep == nil then
		sep = ""
	end
	if arr ~= nil then
		for i,v in ipairs(arr) do
			if str == "" then
				str = v
			else
				str = str .. sep ..  v
			end
		end
	end
	return str
end
