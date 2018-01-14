

(async function() {
    const instance = await phantom.create();
    const page = await instance.createPage();
    const status = await page.open('https://www.cryptopia.co.nz/Exchange?market=SMART_BTC');
    const content = await page.property('content');
    fs.writeFile('foo.html',content);
    await instance.exit();
}());
