import Vue from 'vue'
import Router from 'vue-router'
import { useAuthStore } from '@/stores/auth';

// Containers
const TheContainer = () =>
    import('@/containers/TheContainer')

// Views - Pages
const Page404 = () =>
    import('@/views/pages/Page404')
const Page500 = () =>
    import('@/views/pages/Page500')
const Login = () =>
    import('@/views/pages/Login')

// Home
const Home = () =>
    import('@/views/Home')

// Bans
const Bans = () =>
    import('@/views/bans/Bans')
const IPBlacklist = () =>
    import('@/views/ipblacklist/IPBlacklist')

const PlayerSearch = () =>
    import('@/views/players/PlayerSearch')

const PlayerDetails = () =>
    import('@/views/players/PlayerDetails')

const Skins = () =>
    import('@/views/skins/Skins')

const Items = () =>
    import('@/views/items/Items')


const AttestationBanInfo = () =>
    import('@/views/meta_attestation/BanInfo')
const AttestationDetail = () =>
    import('@/views/meta_attestation/ListByPlatformID')

// Cloud
const CloudInfo = () =>
    import('@/views/cloud/CloudInfo')

Vue.use(Router)

const rtr = new Router({
    mode: 'hash', // https://router.vuejs.org/api/#mode
    linkActiveClass: 'active',
    scrollBehavior: () => ({ y: 0 }),
    routes: configRoutes()
})

rtr.beforeEach(async (to, from, next) => {
    if (to.matched.some(record => record.meta.skipAuth)) {
        return next();
    }

    const authStore = useAuthStore();

    const isAuthenticated = await authStore.ensureAuthenticated();
    if (isAuthenticated) {
        return next();
    }

    return next({ "name": "Login" });
});

export default rtr;

function configRoutes() {
    return [{
        path: '/',
        redirect: '/dashboard',
        name: 'Home',
        component: TheContainer,
        children: [{ path: 'diagnostics', name: 'Diagnostics', component: () => import('@/views/diagnostics/Diagnostics.vue') }, {
            path: 'dashboard',
            name: 'Dashboard',
            component: Home
        },
        {
            path: 'cloud_service',
            redirect: '/cloud_service/info',
            name: 'Cloud Service',
            component: {
                render(c) { return c('router-view') }
            },
            children: [
                { path: 'info', name: 'Cloud Info', component: CloudInfo },
            ]
        },
        {
            path: 'skins',
            redirect: '/skins/all',
            component: {
                render(c) { return c('router-view') }
            },
            children: [{
                path: 'all',
                name: 'Skins',
                component: Skins
            }]
        },
        {
            path: 'items',
            redirect: '/items/all',
            component: {
                render(c) { return c('router-view') }
            },
            children: [{
                path: 'all',
                name: 'Items',
                component: Items
            }]
        },
        {
            path: 'meta-attestation',
            redirect: { name: 'BanInfo' }, // Redirecting by NAME is safer
            name: 'MetaQuestAttestation',
            component: {
                render(c) { return c('router-view') }
            },
            children: [
                { path: 'baninfo', name: 'BanInfo', component: AttestationBanInfo },
                { path: 'view/:platform_id', name: 'PlatformId-uniqueIdDetail', component: AttestationDetail, props: true },
            ]
        },
        {
            path: 'player',
            redirect: '/player/search',
            component: {
                render(c) { return c('router-view') }
            },
            children: [{
                path: 'search',
                name: 'PlayerSearch',
                component: PlayerSearch
            }, {
                path: 'details',
                name: 'PlayerDetails',
                component: PlayerDetails
            }]
        },
        {
            path: 'bans',
            redirect: '/bans/all',
            component: {
                render(c) { return c('router-view') }
            },
            children: [{
                path: 'all',
                name: 'Bans',
                component: Bans
            }]
        },
        {
            path: 'ipblacklist',
            redirect: '/ipblacklist/all',
            name: 'IPBlacklist',
            component: {
                render(c) { return c('router-view') }
            },
            children: [{
                path: 'all',
                name: 'BlockedIPlist',
                component: IPBlacklist
            }]
        }
        ]
    },
    {
        path: '/pages',
        redirect: '/pages/404',
        name: 'Pages',
        component: {
            render(c) { return c('router-view') }
        },
        children: [{
            path: '404',
            name: 'Page404',
            meta: { skipAuth: true },
            component: Page404
        },
        {
            path: '500',
            name: 'Page500',
            meta: { skipAuth: true },
            component: Page500
        },
        {
            path: 'login',
            name: 'Login',
            meta: { skipAuth: true },
            component: Login
        }
        ]
    }
    ]
}
