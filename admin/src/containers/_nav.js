export default [{
    _name: 'CSidebarNav',
    _children: [
        {
            _name: 'CSidebarNavTitle',
            _children: ['Dashboard']
        },
        {
            _name: 'CSidebarNavItem',
            name: 'Players',
            to: '/player/search',
            icon: 'cil-walk'
        },
        {
            _name: 'CSidebarNavItem',
            name: 'Skins',
            to: '/skins/all',
            icon: 'cil-star'
        },
        {
            _name: 'CSidebarNavItem',
            name: 'Items',
            to: '/items/all',
            icon: 'cil-basket'
        },
        {
            _name: 'CSidebarNavItem',
            name: 'Diagnostics',
            to: '/diagnostics',
            icon: 'cil-list'
        },
        {
            _name: 'CSidebarNavItem',
            name: 'Active Game Bans',
            to: '/bans/all',
            icon: 'cil-ban'
        },
        {
            _name: 'CSidebarNavItem',
            name: 'IP Blacklist',
            to: '/ipblacklist/all',
            icon: 'cil-ban'
        },
        {
            _name: 'CSidebarNavItem',
            enable_attestation: true,
            name: 'Meta Attestation',
            to: '/meta-attestation/baninfo',
            icon: 'cil-ban'
        },
        {
            _name: 'CSidebarNavItem',
            name: 'Cloud Info',
            to: '/cloud_service/info',
            icon: 'cil-cloud'
        },
    ]
}]
