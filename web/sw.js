if(!self.define){let e,s={};const i=(i,c)=>(i=new URL(i+".js",c).href,s[i]||new Promise((s=>{if("document"in self){const e=document.createElement("script");e.src=i,e.onload=s,document.head.appendChild(e)}else e=i,importScripts(i),s()})).then((()=>{let e=s[i];if(!e)throw new Error(`Module ${i} didn’t register its module`);return e})));self.define=(c,n)=>{const a=e||("document"in self?document.currentScript.src:"")||location.href;if(s[a])return;let r={};const o=e=>i(e,a),f={module:{uri:a},exports:r,require:o};s[a]=Promise.all(c.map((e=>f[e]||o(e)))).then((e=>(n(...e),r)))}}define(["./workbox-118fddf1"],(function(e){"use strict";self.skipWaiting(),e.clientsClaim(),e.precacheAndRoute([{url:"assets/index-0e8bf8a6.css",revision:null},{url:"assets/index-2128d00b.css",revision:null},{url:"assets/index-2db47a63.js",revision:null},{url:"assets/index-c6f02c8b.js",revision:null},{url:"assets/setting-c6ca7b14.svg",revision:null},{url:"compose.html",revision:"2c3f93033c3f4cef8136ff5a993a087b"},{url:"favicon.ico",revision:"1272c70e1b86b8956598a0349d2f193c"},{url:"img/compose.svg",revision:"4242b76bb8f4da0baf7a75edab0c6754"},{url:"img/logo.svg",revision:"1da58864f14c1a8c28f8587d6dcbc5d0"},{url:"img/pwa/logo-192.png",revision:"be40443731d9d4ead5e9b1f1a6070135"},{url:"img/pwa/logo-512.png",revision:"1217f1c90acb9f231e3135fa44af7efc"},{url:"index.html",revision:"3446c007832ca09298c737ded8c8ac9c"},{url:"js/bing/chat/amd.js",revision:"8d773dc8f2e78b9d29e990aed7821774"},{url:"js/bing/chat/config.js",revision:"3bd7b84479a1f1dcc850abdd4d383a3c"},{url:"js/bing/chat/core.js",revision:"8c11521fd9f049b6ac91e5ad415c2db1"},{url:"js/bing/chat/global.js",revision:"2b5db148d13525a415ecf4e2c929ec43"},{url:"js/bing/chat/lib.js",revision:"1a0f8f43cc025b7b5995e885fed1a3e6"},{url:"registerSW.js",revision:"bf6c2f29aef95e09b1f72cf59f427a55"},{url:"./img/pwa/logo-192.png",revision:"be40443731d9d4ead5e9b1f1a6070135"},{url:"./img/pwa/logo-512.png",revision:"1217f1c90acb9f231e3135fa44af7efc"},{url:"manifest.webmanifest",revision:"ae4ef030ae5d2d4894669fd82aac028d"}],{}),e.cleanupOutdatedCaches(),e.registerRoute(new e.NavigationRoute(e.createHandlerBoundToURL("index.html"))),e.registerRoute(/(.*?)\.(js|css|ts)/,new e.CacheFirst({cacheName:"js-css-cache",plugins:[new e.ExpirationPlugin({maxEntries:100,maxAgeSeconds:604800}),new e.CacheableResponsePlugin({statuses:[0,200]})]}),"GET"),e.registerRoute(/(.*?)\.(png|jpe?g|svg|gif|bmp|psd|tiff|tga|eps|ico)/,new e.CacheFirst({cacheName:"image-cache",plugins:[new e.ExpirationPlugin({maxEntries:100,maxAgeSeconds:604800}),new e.CacheableResponsePlugin({statuses:[0,200]})]}),"GET")}));
