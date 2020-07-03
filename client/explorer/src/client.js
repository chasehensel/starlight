const basePath = "https://e329e49c8232.ngrok.io/";
const methods = ["create", "findOne", "findMany", "update", "updateMany"];

async function call(uri, params) {
  const res = await fetch(uri, {
    method: "POST",
    body: JSON.stringify(params)
  });
  const responseBody = await res.json();
  if ("code" in responseBody) {
    return Promise.reject(responseBody);
  }
  return responseBody.data;
}

var client = {
  api: new Proxy(
    {},
    {
      get: function(target, resource) {
        var out = {};
        methods.forEach(method => {
          out[method] = params => {
            return call(basePath + "api/" + resource + "." + method, params);
          };
        });
        return out;
      }
    }
  ),
  views: {
    rpc: new Proxy(
      {},
      {
        get: function(target, resource) {
          return params => {
            return call(basePath + "views/rpc/" + resource, params);
          };
        }
      }
    ),
    repl: params => {
      return call(basePath + "views/repl", params);
    }
  },
  log: params => {
    return call(basePath + "log.scan", params);
  }
};

export default client;