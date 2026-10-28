<template>
    <div>
        
        <CContainer>
            <CHeader class="mb-3 p-3">
                <h1>Dashboard</h1>
            </CHeader>
            
            <CRow>
                <CCol>
                    Hi!<br /><br /><br />
                </CCol>
			</CRow>			
			<CRow>
				<CCol>				
					Photon authenticate {{ this.photonAuthenticateUrl }}
					<CEmbed tag="div" type="embed" aspect="21by9" :src="this.photonAuthenticateUrlWithParam" frameborder="0" allow="accelerometer;autoplay; encrypted-media; gyroscope; picture-in-picture" />
					<br/>
				</CCol>
			</CRow>			
			<CRow>			
				<CCol>
					<CTabs variant="pills" :active-tab="0">
						<CTab title="Useful Links">
							<CListGroup>
								<CListGroupItem href="https://github.com/justinliebregts?tab=repositories" color="info" target="_blank">
                  GitHub
								</CListGroupItem>
								<CListGroupItem href="https://dash.cloudflare.com/login" target="_blank">
                  CLOUDFLARE
								</CListGroupItem>
								<CListGroupItem href="https://www.tencentcloud.com/account/login/subAccount" color="primary" target="_blank">
                  TencentCloud CAM user login
								</CListGroupItem>
							</CListGroup>
						</CTab>
					</CTabs>
				</CCol>
			</CRow>
		</CContainer>
    </div>
</template>

<script>
export default {
    name: "Home",
  data (){
    return {
			photonAuthenticateUrl: this.$gstate.getDefaultUrl() +'/photon/authenticate',
			photonAuthenticateUrlWithParam: this.$gstate.getDefaultUrl() +'/photon/authenticate?usr=3724467550970486&pla=OCULUS&tkn=Xe94YkdkD9ejWG5rp42nugeHnEr&secret=128fd21c06e35ef8d5d6187c5c452cd7561035b4',
        };
    },
};
</script>
