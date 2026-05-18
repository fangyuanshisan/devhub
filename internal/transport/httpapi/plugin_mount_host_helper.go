package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// v1.8.2: shared official-plugin iframe host helper.
// It is served by the backend so both Astro pages and Go SEO pages can reuse it
// without copying large inline scripts. It must never load remote JS.
func (s *Server) pluginMountHostHelperJS(c *gin.Context) {
	c.Header("Content-Type", "application/javascript; charset=utf-8")
	c.String(http.StatusOK, pluginMountHostHelperJS())
}

func pluginMountHostHelperJS() string {
	// Keep this helper small and safe:
	// - Only supports official built-in plugins allowlist (phase 1: official_announcement).
	// - No remote URL support.
	// - Enforces strict postMessage checks (source/origin/plugin_code/mount_id/type).
	// - Never exposes any token/secret to iframe.
	return `(function(){
  'use strict';
  if (window.DevHubOfficialPluginMountHost) return;

  var SCHEMA_VERSION = '1';
  var ALLOWED_MESSAGE_TYPES = new Set(['devhub.plugin.ready','devhub.plugin.config.read','devhub.plugin.audit.write']);
  var PLUGINS = {
    official_announcement: {
      code: 'official_announcement',
      iframePath: '/plugins/official-announcement/iframe',
      contextAPI: '/api/v1/plugins/official-announcement/context',
      auditAPI: '/api/v1/plugins/official-announcement/audit-events',
      iframeTitle: 'Official Announcement'
    }
  };

  function uid(prefix){
    return String(prefix || 'id') + '_' + Math.random().toString(16).slice(2) + Date.now().toString(16);
  }

  function getPluginCode(el){
    if (!el) return '';
    if (el.dataset && el.dataset.pluginCode) return String(el.dataset.pluginCode || '').trim();
    // Backward compat: v1.8.1 used data-official-announcement-host
    if (el.hasAttribute && el.hasAttribute('data-official-announcement-host')) return 'official_announcement';
    return '';
  }

  function getArea(el){
    var area = el && el.dataset ? String(el.dataset.area || '').trim() : '';
    if (area === 'admin') return 'admin';
    return 'frontend';
  }

  function getCommunitySlug(el){
    return el && el.dataset ? String(el.dataset.communitySlug || '').trim() : '';
  }

  function safeString(v){ return String(v == null ? '' : v); }

  function adminHeaders(){
    try {
      var token = window.sessionStorage && window.sessionStorage.getItem('devhub_admin_token');
      if (token) return { Authorization: 'Bearer ' + token };
    } catch (e) {}
    return {};
  }

  function shouldShowByConfig(cfg){
    cfg = cfg || {};
    if (!cfg.enabled) return false;
    if (!safeString(cfg.message).trim()) return false;
    return true;
  }

  function postToIframe(iframe, message){
    try {
      if (!iframe || !iframe.contentWindow) return;
      // The iframe is intentionally sandboxed without allow-same-origin, so its
      // origin is opaque ("null"). targetOrigin must be "*" for the browser to
      // deliver the message; inbound messages are still bound to contentWindow,
      // mount_id, plugin_code, schema_version, and message type below.
      iframe.contentWindow.postMessage(message, '*');
    } catch (e) {}
  }

  function mountOne(el, opts){
    opts = opts || {};
    if (!el) return;
    if (el.dataset && el.dataset.mounted === '1') return;
    if (typeof el.__devhubPluginUnmount === 'function') {
      try { el.__devhubPluginUnmount(); } catch (e) {}
    }

    var pluginCode = opts.pluginCode || getPluginCode(el);
    var plugin = PLUGINS[pluginCode];
    if (!plugin) return;

    var area = opts.area || getArea(el);
    var communitySlug = (opts.communitySlug != null) ? String(opts.communitySlug) : getCommunitySlug(el);

    var mountId = (el.dataset && el.dataset.mountId) ? String(el.dataset.mountId).trim() : '';
    if (!mountId) {
      mountId = uid('mnt');
      if (el.dataset) el.dataset.mountId = mountId;
    }
    var requestId = uid('req');

    // Fetch context/config from browser-safe Host API.
    var url = new URL(plugin.contextAPI, window.location.origin);
    url.searchParams.set('mount_id', mountId);
    url.searchParams.set('area', area);
    if (communitySlug) url.searchParams.set('community_slug', communitySlug);

    fetch(url.toString(), { credentials: (area === 'admin') ? 'include' : 'same-origin', headers: (area === 'admin') ? adminHeaders() : {} })
      .then(function(res){ return res && res.ok ? res.json() : null; })
      .then(function(data){
        if (!data) return;
        if (data.visible === false) return;
        var cfg = data.config || {};
        if (!shouldShowByConfig(cfg)) return;

        // Create iframe (built-in route only; never remote).
        var iframe = document.createElement('iframe');
        var reply = function(type, payload){
          postToIframe(iframe, {
            schema_version: SCHEMA_VERSION,
            type: type,
            plugin_code: pluginCode,
            mount_id: mountId,
            request_id: requestId,
            payload: payload || {}
          });
        };

        var audit = function(action, metadata){
          fetch(plugin.auditAPI, {
            method: 'POST',
            credentials: (area === 'admin') ? 'include' : 'same-origin',
            headers: Object.assign({ 'Content-Type': 'application/json' }, (area === 'admin') ? adminHeaders() : {}),
            body: JSON.stringify({
              mount_id: mountId,
              area: area,
              community_slug: communitySlug || undefined,
              request_id: requestId,
              action: action,
              metadata: metadata || {}
            })
          }).catch(function(){});
        };

        var onMessage = function(event){
          if (!event || (event.origin !== window.location.origin && event.origin !== 'null')) return;
          if (event.source !== iframe.contentWindow) return;
          var msg = event.data;
          if (!msg || msg.schema_version !== SCHEMA_VERSION) return;
          if (msg.plugin_code !== pluginCode) return;
          if (msg.mount_id !== mountId) return;
          if (!ALLOWED_MESSAGE_TYPES.has(msg.type)) return;

          if (msg.type === 'devhub.plugin.ready') {
            reply('devhub.plugin.context', { context: data.context || {}, config: cfg });
            return;
          }
          if (msg.type === 'devhub.plugin.config.read') {
            reply('devhub.plugin.config.result', { config: cfg });
            return;
          }
          if (msg.type === 'devhub.plugin.audit.write') {
            var payload = msg.payload || {};
            var act = safeString(payload.action).trim();
            if (act && act.indexOf(pluginCode + '.') === 0) {
              audit(act, payload.metadata || {});
            }
            reply('devhub.plugin.audit.result', { ok: true });
            return;
          }
        };

        window.addEventListener('message', onMessage);
        iframe.addEventListener('load', function(){
          reply('devhub.plugin.context', { context: data.context || {}, config: cfg });
        });
        var iframeURL = new URL(plugin.iframePath, window.location.origin);
        iframeURL.searchParams.set('mount_id', mountId);
        iframeURL.searchParams.set('area', area);
        iframe.src = iframeURL.pathname + '?' + iframeURL.searchParams.toString();
        iframe.sandbox = 'allow-scripts';
        iframe.referrerPolicy = 'no-referrer';
        iframe.style.width = '100%';
        iframe.style.border = '0';
        iframe.style.height = '56px';
        iframe.setAttribute('title', plugin.iframeTitle || plugin.code);
        el.innerHTML = '';
        el.appendChild(iframe);

        if (el.dataset) el.dataset.mounted = '1';
        el.__devhubPluginUnmount = function(){
          window.removeEventListener('message', onMessage);
          if (el.dataset) el.dataset.mounted = '0';
          delete el.__devhubPluginUnmount;
        };
      })
      .catch(function(){});
  }

  function mountAll(){
    // New generic attribute (v1.8.2)
    document.querySelectorAll('[data-devhub-plugin-mount]').forEach(function(el){
      mountOne(el, {});
    });
    // Backward compat (v1.8.1)
    document.querySelectorAll('[data-official-announcement-host]').forEach(function(el){
      mountOne(el, { pluginCode: 'official_announcement' });
    });
  }

  window.DevHubOfficialPluginMountHost = { mount: mountOne, mountAll: mountAll };

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', mountAll);
  } else {
    mountAll();
  }
})();`
}
