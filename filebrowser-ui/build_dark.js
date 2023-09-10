import { readFile, writeFileSync } from 'fs';

const regex = /<meta\sclass="dark-theme">/gm;

readFile('./dist/index.html', 'utf8', (err, data) => {
    if (err) {
        console.error(err);
        return;
    }
    data = data.replace(regex, "<link rel=\"stylesheet\" href=\"themes/dark.css\">");
    writeFileSync("./dist/index-dark.html", data)
});