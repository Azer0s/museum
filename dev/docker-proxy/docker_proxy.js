import express from 'express';

let app = express();

app.use('/proxy/:container/:port/*', async (req, res) => {
    let actualPath = req.params[0];
    let container = req.params.container;
    let port = req.params.port;

    let url = `http://${container}:${port}/${actualPath}`;

    console.log(`Proxying ${req.method} ${url}`);

    let response = await fetch(url, {
        method: req.method,
        headers: req.headers,
        body: req.body
    });

    res.status(response.status);
    res.set(response.headers);
    res.send(await response.text());
});

app.listen(3000, () => {
    console.log('Listening on port 3000');
});