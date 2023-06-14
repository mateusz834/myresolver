document.addEventListener("DOMContentLoaded", () => {
  const addrElement = document.getElementById("addr-location");
  const queryDomain = randomHex() + "." + "rand.api.get." + location.hostname + ".";
  setTimeout(async () => {
    try {
      await fetch(location.protocol + "//" + queryDomain);
    } catch { }
  }, 0);
  setTimeout(async () => {
    for (let i = 0; i < 10; i++) {
      const res = await fetch("/api/who-resolved?domain=" + queryDomain);
      if (res.status == 200) {
        const json = await res.json();
        addrElement.innerText = json.addr;
        break;
      }
      await new Promise((resolve) => setTimeout(() => resolve(), 200 * i));
    }
  }, 250);
});


/**
  * @returns {string}
  */
function randomHex() {
  return toHex(crypto.getRandomValues(new Uint8Array(30)));
}


/**
  * @argument {Uint8Array} a
  * @returns {string}
  */
function toHex(a) {
  return a.reduce(
    (prev, val) => {
      var num = val.toString(16);
      if (num.length === 1) num = "0" + num
      return prev + num;
    },
    "",
  )
}
