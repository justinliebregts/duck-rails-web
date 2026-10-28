<template>
    <CFooter :fixed="false">
        <div>
            <small>Server Time: <b>{{ currentServerTime }}</b> | 
                    {{ $gstate.getBaseUrl() }} | 
                    v.{{ $gstate.getVersion() }}</small>
        </div>
        <div class="mfs-auto">
            <span class="mr-1" target="_blank">Powered by lebo</span>
        </div>
    </CFooter>
</template>

<script>
export default {
    name: "TheFooter",
    computed: {
        currentServerTime(){
            return new Date(new Date().toUTCString().slice(0, -3));
        }
    }
};
</script>
