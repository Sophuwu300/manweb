var style = null;

const ValidKeys = ["theme", "contrast", "scale"];
const Valid_theme = ["dark", "warm", "light"];
const Valid_contrast = ["low", "medium", "high"];
const Valid_scale = ["-", "0", "+"];


function SetStyle() {
    let tmp = ":root {\n";
    ValidKeys.forEach(key => {
        let value = localStorage.getItem(key);
        if (value) tmp += "--" + key + ": " + value + ";\n";
    });
    tmp += "}\n\n";
    style.innerHTML = tmp;
    ValidKeys.forEach(key => function (key) {
        if (butts[key]["arr"].length < 3) return;
        let value = localStorage.getItem(key);
        if (!value) {
            if (key == "theme") value = "dark";
            else if (key == "contrast") value = 1;
            else return;
        }
        if (value >= 0 && value <= 2) value = Valid_contrast[value];
        else if (value.startsWith("var(--")) value = value.substring("var(--".length, value.length - 1);
        butts[key]["arr"].forEach(b => b.classList.remove("butt-set"));
        butts[key]["map"][value].classList.add("butt-set");
    }(key));
}
function SaveValue(key, value) {
    if (!ValidKeys.includes(key)) return;
    localStorage.setItem(key, value);
    SetStyle();
}


function SetTheme(theme=null) {
    if (theme==null) {
        let value = localStorage.getItem("theme");
        if (!value) value = "var(--dark)";
        value = value.replace("var(--", "").replace(")", "");
        HlButt("theme", value);
    }
    if (!Valid_theme.includes(theme)) return;
    theme = "var(--" + theme + ")";
    SaveValue("theme", theme);
}

function SetContrast(contrast=null) {
    if (contrast==null) {

        if (!value) value = "1";
        value = Valid_contrast[value];
        HlButt("contrast", value);
    }
    if (!Valid_contrast.includes(contrast)) return;
    SaveValue("contrast", Valid_contrast.indexOf(contrast));
}

function SetScale(scale) {
    if (!Valid_scale.includes(scale)) return;
    let value = localStorage.getItem("scale");
    if (!value) value = 1;
    let diff = (0.1 * (Valid_scale.indexOf(scale) - 1));
    if (diff != 0) {
        value = parseFloat(value) + diff;
        if (value < 0.1) value = 0.1;
        if (value > 2) value = 2;
        value = value.toFixed(1);
        SaveValue("scale", value);
    }
    document.getElementById("scaleVal").innerHTML = (value*100).toFixed(0) + "%"
}

function ResetStyle() {
    ValidKeys.forEach(key => {
        localStorage.removeItem(key);
    });
    SetStyle();
    SetScale("0");
}
function H(elem) {
    return {l: elem.innerHTML.toLowerCase(), n: elem.innerHTML};
}

function Menu(wants="search"){
    let mainClick = [`document.getElementById("main-content").`,`EventListener("click", function () {Menu()})`];

    let n = document.getElementById(wants);
    if (!n || !n.classList.contains("menu")) return;

    eval(mainClick.join(function (b){
        if (b) return "remove";
        return "add";
    }(wants == "search")));

    let l = function (elem) {return {
        e: function (e){e==elem},
        h: elem.classList.contains("hidden"),
        t: function () {elem.classList.toggle("hidden")},
    };};
    document.querySelectorAll(".menu").forEach(ee => function (e){
        if (e.e(n) || e.h) return;
        e.t();
    }(l(ee)));
    n = l(n);
    if (n.h) n.t();
}

function Butt() {
    let butts = new Object();
    ValidKeys.forEach(key => function (key) {
        butts[key]=new Object();
        eval("Valid_"+key).forEach(v => function (v) {
            butts[key][v]=null;
        }(v));
    }(key));
    return butts;
};

var butts = Object();

function SetButts() {
    document.querySelectorAll("#settings > div > h3").forEach(elemH => function (elemH,n) {
    if (!ValidKeys.includes(n.l)) return;
    butts[n.l]= new Object({});
    butts[n.l]["h"] = elemH;
    butts[n.l]["div"] = elemH.parentElement;
    butts[n.l]["arr"] = [];
    butts[n.l]["map"] = new Object({});
    elemH.parentElement.querySelectorAll("button").forEach(butt => function (butt, fun, v) {
        let i = butts[n.l]["arr"].push(butt);
        butts[n.l]["map"][v] = butts[n.l]["arr"][i-1];
        butt.addEventListener("click",  function () {fun(v)});
    }(butt,eval("Set" + n.n), H(butt).l));
    }(elemH,H(elemH)));
}

function ChangeRawQuery(s="") {

    let u = document.URL;
    let i = u.indexOf("?");
    if (s.length > 0 && !s.startsWith("?")) {
        s = "?" + s;
    }
    if (i == -1) {
        return u + s;
    }
    return u.substring(0, i) + s;
}
function SetRawQuery(s="") {
    let u = ChangeRawQuery(s);
    window.history.pushState({"html":document.toString(),"pageTitle":document.title},"", u);
}
function GoToRawQuery(s) {
    let u = ChangeRawQuery(s);
    window.location.href = u;
}
function index(){
  let a = ""
	let q = document.querySelectorAll(".Sh")
  q.forEach(e=>{
  if(e.id!="")
    a+=`<a href="#`+e.id+`">`+e.innerText+`</a>\n`
    //console.log(e.id, e.innerText)
})
  a=`<section style="display: flex;
    flex-direction: column;" id="index" class="Sh"><h1>INDEX</h1>`+a+`</section>`
  console.log(a)
  let h = document.getElementById("index")
  h.outerHTML = a
}
function makeIndex (){
document.querySelector("#NAME").parentElement.outerHTML+=`<p><a onclick="index()" id="index">Create Index</a></p>`
let e=document.querySelector("#index")
}

document.addEventListener("DOMContentLoaded", function () {
    style = document.getElementById("styleCss");
    SetButts();
    SetStyle();
    SetScale("0");
    makeIndex();
});