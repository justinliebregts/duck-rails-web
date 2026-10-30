import { shallowMount } from '@vue/test-utils'
import Diagnostics from '@/views/diagnostics/Diagnostics.vue'
import Login from '@/views/pages/Login.vue'

const tick = () => new Promise(resolve => setTimeout(resolve, 0))
function setup(component = Diagnostics) {
  const axios = { get: jest.fn().mockResolvedValue({ data: { data: { reports: [], total: 0 } } }), post: jest.fn().mockResolvedValue({ data: { success: true } }) }
  const auth = { apiOptions: {}, apiJsonOptions: {}, tmpLoginForever: jest.fn() }
  const router = { replace: jest.fn() }
  return { wrapper: shallowMount(component, { mocks: { axios, $auth: auth, $gstate: { defaultUrl: 'https://example.test/api/v1' }, $router: router } }), axios, auth, router }
}
test('loads filters and offset and never presents a failed query as an empty inbox', async () => {
  const { wrapper, axios } = setup(); await tick()
  wrapper.setData({ status: 'resolved', offset: 100 })
  axios.get.mockResolvedValueOnce({ data: { data: { reports: [], total: 101 } } })
  await wrapper.vm.load()
  expect(axios.get.mock.calls[1][1].params).toEqual({ status: 'resolved', offset: 100 })
  axios.get.mockRejectedValueOnce({ response: { status: 401 } }); await wrapper.vm.load(); await tick()
  expect(wrapper.text()).toContain('Admin access rejected')
  expect(wrapper.text()).not.toContain('No resolved reports.')
})
test('renders untrusted report text safely and tolerates malformed payload', async () => {
  const { wrapper, axios } = setup(); await tick()
  axios.get.mockResolvedValueOnce({ data: { data: { id: 'abc', summary: '<img src=x onerror=alert(1)>', status: 'open', payload: '{broken' } } })
  await wrapper.vm.open('abc'); await tick()
  expect(wrapper.find('img').exists()).toBe(false)
  expect(wrapper.text()).toContain('<img src=x onerror=alert(1)>')
  expect(wrapper.vm.payloadError).toBe(true)
})
test('requires resolution evidence and updates status only after success', async () => {
  const { wrapper, axios } = setup(); await tick()
  wrapper.setData({ selected: { id: 'abc', status: 'open' }, resolution: '  ' })
  await wrapper.vm.changeStatus(); expect(axios.post).not.toHaveBeenCalled()
  wrapper.setData({ resolution: 'Verified fix' }); axios.post.mockRejectedValueOnce(new Error())
  await wrapper.vm.changeStatus(); expect(wrapper.vm.selected.status).toBe('open')
  await wrapper.vm.changeStatus(); expect(wrapper.vm.selected.status).toBe('resolved')
  await wrapper.vm.changeStatus(); expect(wrapper.vm.selected.status).toBe('open')
})
test('login validates token before storing and does not leak server errors', async () => {
  const { wrapper, axios, auth, router } = setup(Login)
  wrapper.setData({ token: 'Bearer test-token' }); axios.get.mockRejectedValueOnce({ response: { status: 401 } })
  await wrapper.vm.login(); expect(auth.tmpLoginForever).not.toHaveBeenCalled()
  await wrapper.vm.login(); expect(auth.tmpLoginForever).toHaveBeenCalledWith('Bearer test-token')
  expect(router.replace).toHaveBeenCalledWith('/diagnostics'); expect(wrapper.vm.token).toBe('')
})
