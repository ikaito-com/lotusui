// gallery-embed.js — lazy-load the lotusui gallery WASM into the docs page.
//
// One Go runtime per page, hosted in #giowindow (moved into the active
// demobox). No iframe: demos share the page's theme pickers and scroll.
(function (global) {
  "use strict";

  function patchGL() {
    if (patchGL.done) return;
    patchGL.done = true;
    const orig = HTMLCanvasElement.prototype.getContext;
    HTMLCanvasElement.prototype.getContext = function (type, attrs) {
      if (type === "webgl" || type === "webgl2") {
        attrs = Object.assign({}, attrs || {}, { preserveDrawingBuffer: true });
      }
      return orig.call(this, type, attrs);
    };
  }

  function absURL(base, file) {
    // Resolve against the document so Root-relative ("../../gallery/")
    // and site-root ("/gallery/") both work from any docs path.
    try {
      return new URL(file, new URL(base, global.location.href)).href;
    } catch (_) {
      return base + file;
    }
  }

  function loadScript(src) {
    return new Promise((resolve, reject) => {
      const existing = document.querySelector('script[src="' + src + '"]');
      if (existing) {
        if (typeof Go !== "undefined") {
          resolve();
          return;
        }
        existing.addEventListener("load", () => resolve());
        existing.addEventListener("error", () =>
          reject(new Error("failed to load " + src))
        );
        return;
      }
      const s = document.createElement("script");
      s.src = src;
      s.onload = () => resolve();
      s.onerror = () => reject(new Error("failed to load " + src));
      document.head.appendChild(s);
    });
  }

  // Create #giowindow parked inside parent (a demobox) BEFORE go.run so
  // Gio sizes to the box — never a full-viewport overlay on <body>.
  function ensureHost(parent) {
    let host = document.getElementById("giowindow");
    if (!host) {
      host = document.createElement("div");
      host.id = "giowindow";
      host.className = "demohost";
      host.setAttribute("aria-hidden", "true");
    } else {
      host.classList.add("demohost");
    }
    if (parent && host.parentNode !== parent) {
      parent.appendChild(host);
    }
    return host;
  }

  function fail(host, err) {
    if (host) {
      host.classList.remove("demohost-loading");
      host.classList.add("demohost-error");
      host.dataset.error = (err && (err.message || String(err))) || "demo failed";
    }
    starting = null;
    throw err;
  }

  let starting = null;

  /**
   * @param {string} base gallery directory URL ending with /
   * @param {HTMLElement} [parkIn] demobox to host #giowindow before go.run
   */
  global.lotusuiStartGallery = function (base, parkIn) {
    if (global.lotusuiGalleryHost) {
      const host = global.lotusuiGalleryHost;
      if (parkIn && host.parentNode !== parkIn) parkIn.appendChild(host);
      return Promise.resolve(host);
    }
    if (starting) return starting;

    starting = (async () => {
      const host = ensureHost(parkIn || document.body);
      host.classList.add("demohost-loading");
      host.classList.remove("demohost-error");
      try {
        patchGL();
        const execURL = absURL(base, "wasm_exec.js");
        const wasmURL = absURL(base, "gallery.wasm");
        await loadScript(execURL);
        if (typeof Go === "undefined") {
          throw new Error("wasm_exec.js did not define Go");
        }
        const go = new Go();
        // arrayBuffer path: instantiateStreaming is picky about MIME and
        // aborts under virtual-time / aggressive navigation; one fetch is enough.
        const buf = await (await fetch(wasmURL)).arrayBuffer();
        const result = await WebAssembly.instantiate(buf, go.importObject);
        host.classList.remove("demohost-loading");
        global.lotusuiGalleryHost = host;
        // go.run never returns — schedule after we expose the host.
        setTimeout(() => {
          try {
            go.run(result.instance);
          } catch (err) {
            console.error("lotusui gallery:", err && err.name, err && err.message, err);
            host.classList.add("demohost-error");
            host.dataset.error = (err && err.message) || String(err);
          }
        }, 0);
        return host;
      } catch (err) {
        console.error("lotusui gallery:", err && err.name, err && err.message, err);
        fail(host, err);
      }
    })();

    return starting;
  };

  global.lotusuiGalleryCanvas = function () {
    const host = document.getElementById("giowindow");
    return host ? host.querySelector("canvas") : null;
  };
})(typeof window !== "undefined" ? window : globalThis);
