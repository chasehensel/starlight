const CoreApi = {
	objects: 
	{
		list: {},
		info: {},
	}
}

class HttpRpcClient {
	constructor(basePath, apiSpec) {
		for (let [resource, methods] of Object.entries(apiSpec)) {
			this[resource] = {};
			for (let [method, _] of Object.entries(methods)) {
				this[resource][method] = async (params) => {
					const res = await fetch(basePath + resource + "." + method, {
						method: "POST",
						body: JSON.stringify(params)
					});
					return await res.json();
				}
			}
		}
	}
}

var client = new HttpRpcClient("https://localhost:8080/api/", CoreApi);

module.exports = client;