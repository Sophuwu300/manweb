    /*
    --theme: <integer 0-5>;
    * sets the theme color, 4 is the default. 0 white, 1-3 add yellow filter. 4 is dark, 5 changes header+link colors.
    --contrast: <bool 0/1>;
    * sets high contrast, 0 is the default.
    --font: string;
    * sets the font, 'JetBrains Mono' will be used from google fonts if not set.
    --font-size: <float><unit>;
    * sets the font size, 18px is the default.
     */

const CSSRx = /^(((\d+(\.\d+)?)(%|p[xtc]|r?em|ex|ch|v(min|max|w|h)|([cm]m)|in))|(((?<x>x{1,3}-)?(small|larg)e?((?<=\k<x>)r)?)|medium|normal))$/;


let funcmap = {
    "--theme": function (value) {
        let v = parseInt(value);
        return (v < 0 || v > 5)
    },
    "--contrast": function (value) {
        let v = parseInt(value);
        return (v < 0 || v > 1)
    },
    "--font": function (value) {
        return (value.length < 1)
    },
    "--font-size": function (value) {
        return !CSSRx.test(value)
    }
};


function SaveValue(key, value) {
    let fn = funcmap[key]
    if (typeof(fn) != "function"||fn(value)) {
        return false;
    }
    localStorage.setItem(key, value);
    return true;
}

function DeleteValue(key) {
    localStorage.removeItem(key);
}

// var style = document.createElement('style');
// style.innerText += document.styleSheets[0].cssRules[0].cssText;
// const styleCss = style.innerText;
// style.id = "styleCss";
// document.head.appendChild(style);

const style = document.getElementById("styleCss");
const styleCss = style.innerText;

function setStyle() {
    let tmp = ":root{\n";
    let ar = ["--theme", "--contrast", "--font", "--font-size"];
    for (let key of ar) {
        let value = localStorage.getItem(key);
        if (value) tmp += `${key}:${value};` + "\n";
    }
    tmp += "}";
    style.innerText = tmp + styleCss;
}

document.addEventListener("DOMContentLoaded", setStyle);
document.addEventListener("DOMContentLoaded", function () {
    document.getElementById("SetButt").addEventListener("click", function () {
        let ar = ["--theme", "--contrast", "--font", "--font-size"];
        for (let key of ar) {
            let value = localStorage.getItem(key);
            if (value) document.getElementById(key.replaceAll("-", "")).value = value;
        }
        document.querySelector("article.settings").classList.toggle("hidden");
    });
    document.getElementById("contrast").addEventListener("mousedown", function () {
            let elem = document.getElementById("contrast");
            if (elem.value == 1) elem.value = 0;
            else elem.value = 1;
            elem.classList.toggle("togii");
            if (SaveValue("--contrast", elem.value))setStyle();
    });
    document.getElementById("theme").addEventListener("input", function () {
        if (SaveValue("--theme", document.getElementById("theme").value))setStyle();
    });
});
