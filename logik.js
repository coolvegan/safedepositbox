function CheckPasswordFields() {
  const passwordMinLength = 4;
  const password1 = document.getElementById("password1").value;
  const password2 = document.getElementById("password2").value;
  return password2 === password1 && password1.length >= passwordMinLength;
}

function GetUserInputPassword() {
  return document.getElementById("password1").value;
}

function GetDecodeViewPassword() {
  return document.getElementById("decryptPassword").value;
}

function ValidatePassword() {
  return CheckPasswordFields();
}

async function DecodeAndShow() {
  const fetchkey = document.getElementById("fetchKey").value;
  const response = await fetch(
    "/data?" + new URLSearchParams({ fetchkey: fetchkey }),
    {
      headers: { "Content-Type": "application/json" },
      method: "GET",
    },
  );
  const text = await response.text();
  let dump = JSON.parse(text);

  let password = GetDecodeViewPassword();
  let salt = hexStringToArrayBuffer(atob(dump.salt));
  let key = await GenerateKey(password, new Uint8Array(salt));
  if (password.length == 0 || salt.length == 0 || key.length == 0) {
    alert("Keine Daten vorhanden.");
    return;
  }

  let result = await decryptData(
    hexStringToArrayBuffer(atob(dump.data)),
    hexStringToArrayBuffer(atob(dump.iv)),
    key,
  );
  document.getElementById("decryptTextarea").value = result;
}

function base64ToZahlenArray(base64String) {
  // Base64 dekodieren zu Uint8Array
  const binaryString = atob(base64String);
  const bytes = new Array(binaryString.length);

  for (let i = 0; i < binaryString.length; i++) {
    bytes[i] = binaryString.charCodeAt(i);
  }

  return Array.from(bytes);
}

async function EncodeAndSend() {
  if (!ValidatePassword()) {
    alert(
      "Passwörter sind nicht gleich. Die Länge muss mindestens 8 Zeichen betragen.",
    );
  }
  const salt = crypto.getRandomValues(new Uint8Array(16));
  const key = await GenerateKey(GetUserInputPassword(), salt);
  const hideme = document.getElementById("hideme").value;
  const encryptedDataAndIv = await encryptData(hideme, key);
  await SendToServer(encryptedDataAndIv, salt);
}

async function SendToServer(encryptedDataAndIv, salt) {
  let data = uint8ArrayToHex(new Uint8Array(encryptedDataAndIv.encryptedData));
  let iv = uint8ArrayToHex(new Uint8Array(encryptedDataAndIv.iv));
  let salt_ = uint8ArrayToHex(new Uint8Array(salt));

  const base64obj = {
    "data": btoa(data),
    "iv": btoa(iv),
    "salt": btoa(salt_),
  };
  appViewState("responseView");
  const response = await fetch("/up", {
    headers: { "Content-Type": "application/json" },
    method: "POST",
    body: JSON.stringify(base64obj),
  });
  const text = await response.text();
  //Todo: Trennung von View und Logik
  let ev = document.getElementById("encryptView");
  ev.style.display = "none";
  document.getElementById("responseKey").innerHTML = "Dein Abholcode: " + text;
  console.log(text);
}

async function GenerateKey(secret, salt) {
  const config = {
    pass: secret, // Passwort
    salt: salt, // Salt für die Ableitung
    time: 5, // Anzahl der Iterationen
    mem: 65536, // Speicherverbrauch in KB (z. B. 64 MB)
    parallelism: 1, // Anzahl der parallelen Threads
    hashLen: 32, // Länge des abgeleiteten Schlüssels (32 Bytes = 256 Bit)
    type: argon2.ArgonType.Argon2id, // Argon2id bietet Schutz gegen GPU-Angriffe
  };

  const result = await argon2.hash(config);
  // Schlüsselableitung aus dem Passwort
  const keyMaterial = hexStringToArrayBuffer(result.hashHex);
  const encryptionKey = await crypto.subtle.importKey(
    "raw",
    keyMaterial,
    { name: "AES-GCM" },
    false,
    [
      "encrypt",
      "decrypt",
    ],
  );
  return encryptionKey;
}

function uint8ArrayToHex(uint8Array) {
  return Array.from(uint8Array)
    .map((byte) => byte.toString(16).padStart(2, "0")) // Jedes Byte in Hex konvertieren
    .join(""); // Die Hex-Werte zu einem String zusammenfügen
}

function hexStringToArrayBuffer(hexString) {
  // remove the leading 0x
  hexString = hexString.replace(/^0x/, "");

  // ensure even number of characters
  if (hexString.length % 2 != 0) {
    console.log(
      "WARNING: expecting an even number of characters in the hexString",
    );
  }

  // check for some non-hex characters
  var bad = hexString.match(/[G-Z\s]/i);
  if (bad) {
    console.log("WARNING: found non-hex characters", bad);
  }

  // split the string into pairs of octets
  var pairs = hexString.match(/[\dA-F]{2}/gi);

  // convert the octets to integers
  var integers = pairs.map(function (s) {
    return parseInt(s, 16);
  });

  var array = new Uint8Array(integers);

  return array.buffer;
}

async function encryptData(data, encryptionKey) {
  const iv = crypto.getRandomValues(new Uint8Array(12)); // IV für AES-GCM (12 Bytes)

  const encryptedData = await crypto.subtle.encrypt(
    {
      name: "AES-GCM",
      iv: iv,
    },
    encryptionKey,
    new TextEncoder().encode(data), // Umwandeln der Daten in Bytes
  );

  return { encryptedData, iv };
}

async function decryptData(encryptedData, iv, encryptionKey) {
  try {
    const decryptedData = await crypto.subtle.decrypt(
      {
        name: "AES-GCM",
        iv: iv,
      },
      encryptionKey,
      encryptedData,
    );

    return new TextDecoder().decode(decryptedData); // Umwandeln zurück in Text
  } catch (error) {
    alert("Das Passwort ist fehlerhaft.");
  }
  return "";
}
