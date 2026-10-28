/**
 * 
 */
var gstate = (function () {

    function redirect(url) {
        window.location.href = url;
    }

    console.log("NODE_ENV:" + process.env.NODE_ENV);
    console.log("VUE_APP_API_URL:" + process.env.VUE_APP_API_URL);
    console.log("VUE_APP_DEFAULT_API_URL:" + process.env.VUE_APP_DEFAULT_API_URL);
    console.log("VUE_APP_APPVERSION_PHOTONVERSION:" + process.env.VUE_APP_APPVERSION_PHOTONVERSION);
    //console.log("VUE_APP_API_URL:"+process.env.VUE_APP_API_URL);

    console.log("VUE_APP_PREFAB_ID:" + process.env.VUE_APP_PREFAB_ID);

    return {
        getVersion: function () {
            return document.documentElement.dataset.buildTimestampUtc;
        },
        getAppPhotonVersion: function () {
            return process.env.VUE_APP_APPVERSION_PHOTONVERSION;
        },
        isProd: function () {
            return process.env.NODE_ENV == "production";
        },
        isAttestationEnabled: function () {
            return process.env.VUE_APP_ENABLE_ATTESTATION === 'true';
        },
        k8sNamespace: process.env.VUE_APP_K8S_NAMESPACE,
        baseUrl: process.env.VUE_APP_API_URL,
        getBaseUrl: function () {
            return process.env.VUE_APP_API_URL;
        },
        defaultUrl: process.env.VUE_APP_DEFAULT_API_URL,
        defaultRoot: new URL(process.env.VUE_APP_DEFAULT_API_URL).origin,
        getDefaultInternalProxyUrl() {
            return this.defaultRoot + '/internal/proxy/admin';
        },
        getDefaultInternalProxyUrlWithVersion(path) {
            let urlstr = this.getDefaultInternalProxyUrl();
            if (path) {
                const parts = Array.isArray(path) ? path : [path];
                const cleanParts = parts
                    .filter(p => typeof p === 'string' && p.trim() !== '')
                    .map(p => p.trim().replace(/^\/+/, '').replace(/\/+$/, ''))
                    .filter(Boolean);
                if (cleanParts.length > 0) {
                    urlstr += '/' + cleanParts.join('/');
                }
            }

            const separator = urlstr.includes('?') ? '&' : '?';
            return urlstr + separator + 'use_version=true';
        },
        getDefaultUrl: function () {
            return process.env.VUE_APP_DEFAULT_API_URL;
        },
        getPrefabID: function () {
            return process.env.VUE_APP_PREFAB_ID;
        },
        numberWithCommas: function (x) {
            return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
        },
        getAdminUrl: function (x) {
            if (this.isProd()) {
                return "/admin/#";
            } else {
                return "/#";
            }
        }
    }
})();
export default gstate;