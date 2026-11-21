import { ref } from 'vue'

const showModal = ref(false)
const initialStep = ref(0)

export function useAiOnboardingModal() {
  function open(step = 0) {
    initialStep.value = step
    showModal.value = true
  }

  function close() {
    showModal.value = false
  }

  return {
    showModal,
    initialStep,
    open,
    close,
  }
}
