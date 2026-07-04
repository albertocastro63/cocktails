function handler(event) {
  var request = event.request;
  var uri = request.uri;
  var prMatch = uri.match(/^(\/pr-\d+)(\/.*)?$/);
  if (prMatch && !/\.[a-zA-Z0-9]+$/.test(uri)) {
    request.uri = prMatch[1] + '/index.html';
  }
  return request;
}
