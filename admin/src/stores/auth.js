import { defineStore } from 'pinia'
import Cookies from 'js-cookie'

export const useAuthStore = defineStore('auth', {
    state: () => ({
        token: Cookies.get('acc_token') || null,
        refToken: Cookies.get('ref_token') || null
    }),
    getters: {
        // This creates your 'options' object automatically
        apiOptions: (state) => ({
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
                'Authorization': state.token || ''
            }
        }),
        apiJsonOptions: (state) => ({
            headers: {
                'Content-Type': 'application/json',
                'Authorization': state.token || ''
            }
        }),
    },
    actions: {
        async validateAccessToken() {
            if (this.token == null || this.token.length <=0){
                return false;
            }

            //TODO: now async check with remote api
            return true;
        },
        async validateRefreshToken() {
            if (this.refToken == null || this.refToken.length <=0){
                return false;
            }


            //TODO: now async check with remote api
            return true;
        },
        async ensureAuthenticated() {

            //console.log('Current Token:', this.token);
            //TODO: unsafe once, should be gone when using actual Login page
            if (!await this.validateAccessToken()) {
                this.tmpLoginForever('Xe94YkdkD9ejWG5rp42nugeHnEr');
                return true;
            }
            //TODO: unsafe once

            if (await this.validateAccessToken()){
              return true;
            }

            if (await this.validateRefreshToken()) {
                return await this.refreshTokens();
            }

            return false;
        },
        async refreshTokens() {
            try {
                const response = await axios.post('/api/auth/refresh', {
                    refresh_token: this.refToken
                });

                this.token = response.data.access_token;
                // this.saveLogin();

                return true;
            } catch (error) {
                this.Logout();
                return false;
            }
        },
        queryWithJsonOptions(params = {}) {
            return {
                params: params,
                headers: this.apiJsonOptions.headers
            };
        },
        tmpLoginForever(newAccToken) { //TODO: should be gone when actual login exist
            this.token = newAccToken

            Cookies.set('acc_token', newAccToken);
        },
        saveLogin(newAccToken, newRefToken) {
        //saveLogin(newAccToken, newRefToken) {
            this.token = newAccToken;
            this.refToken = newRefToken;

            Cookies.set('acc_token', newAccToken, { expires: 15 / 1440 });
            Cookies.set('ref_token', newRefToken, { expires: 120 / 1440 });
        },
        Logout(){
            // Clear the LIVE state
            this.token = null;
            this.refToken = null;

            // Clear the permanent storage
            Cookies.remove('acc_token');
            Cookies.remove('ref_token');
        }
    }
})