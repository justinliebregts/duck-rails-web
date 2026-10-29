<template>
    <CSidebar
        fixed
        :hideOnMobileClick="true"
        size="sm"
        :show="true"
        breakpoint="sm"
        :narrow="true"
        :minimize="isMinimized"
        class="sidebar-classic"
    >
        <CSidebarBrand class="d-md-down-none" to="/">
            <span class="p-3 font-weight-bold">Duck Rails</span>
        </CSidebarBrand>
        
        <!--CRenderFunction flat :content-to-render="$options.nav" /-->
        <CRenderFunction flat :content-to-render="filteredNav" />

        <CContainer>
            <CRow class="text-muted sidebar-info pb-3">
                <CCol>
                    <CButton
                            color="light"
                            size="sm"
                            variant="ghost"
                            v-on:click="ToggleMinimize()"
                            class="minimize-button mt-3"
                        >
                        {{minimizeLabel}}
                        </CButton>
                </CCol>
            </CRow>
        </CContainer>
    </CSidebar>
</template>

<script>
import nav from "./_nav";

export default {
    name: "TheSidebar",
    nav,
    data (){
        return {
            isMinimized: false,
        };
    },
    computed: {
      filteredNav() {
        const rawNav = this.$options.nav;
        const isAttestationEnabled = this.$gstate.isAttestationEnabled();

        return rawNav.map(navWrapper => {
          return {
            ...navWrapper,
            _children: navWrapper._children.filter(item => {
              if (item.enable_attestation) {
                return isAttestationEnabled;
              }
              return true;
            })
          };
        });
      },
      minimizeLabel() {
        if(this.isMinimized){
          return ">";
        } else {
          return "<";
        }
        },
      currentServerTime(){
        return new Date(new Date().toUTCString().slice(0, -3));
      }
      },
    methods: {
        ToggleMinimize: function() {
            this.isMinimized = !this.isMinimized;
        }

    }
    
};
</script>
<style>



</style>