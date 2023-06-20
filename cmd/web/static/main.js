document.addEventListener("DOMContentLoaded", () => {
  const queryDomain = randomHex() + "." + "rand.api.get." + location.hostname
  setTimeout(async () => {
    try {
      await fetch(location.protocol + "//" + queryDomain);
    } catch { }
  }, 0);
  setTimeout(async () => {
    for (let i = 0; i < 10; i++) {
      const res = await fetch("/api/who-resolved?domain=" + encodeURIComponent(queryDomain) + ".");
      if (res.status == 200) {
        const json = await res.json();
        const addr = json.addr;
        const asn = json.asn;
        const desc = json.desc;

        document.getElementById("addr-location").innerText = addr;
        document.getElementById("query-domain").innerText = queryDomain;
        if (asn && desc) {
          const v = `ASN ${asn}: '${desc}'`;
          document.getElementById("cli-example").innerText += `\n"Query resolved by: '${addr}'"\n"${v}"`;
          document.getElementById("asn-details").innerText = v;
        } else {
          document.getElementById("cli-example").innerText += `\n"Query resolved by: '${addr}'"`;
        }
        document.getElementById("detecting").classList.add("hidden");
        document.getElementById("detected").classList.remove("hidden");
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
  return toHex(crypto.getRandomValues(new Uint8Array(20)));
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
