<script setup>
import { ref, onMounted } from 'vue'

const show = ref(false)

onMounted(() => {
  if (!localStorage.getItem('grtblog-v2-announcement-closed')) {
    show.value = true
    document.documentElement.classList.add('has-announcement')
  }
})

function close() {
  show.value = false
  localStorage.setItem('grtblog-v2-announcement-closed', 'true')
  document.documentElement.classList.remove('has-announcement')
}
</script>

<template>
  <Transition name="fade">
    <div v-if="show" class="announcement-bar">
      <div class="container">
        <div class="content">
          <span class="badge">New</span>
          <span class="text">GrtBlog v2 开发接近尾声，已发布正式版本，将替代 v1 提供更出色的体验！</span>
          <a href="https://github.com/grtsinry43/grtblog" target="_blank" class="link">了解更多 &rarr;</a>
        </div>
        <button class="close" @click="close" aria-label="Close">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="icon">
            <path d="M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z" />
          </svg>
        </button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.announcement-bar {
  background: var(--vp-c-brand-1);
  color: white;
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  z-index: 1000;
  transition: background-color 0.3s;
}

:deep(.dark) .announcement-bar {
  background: var(--vp-c-brand-3);
}

.container {
  max-width: var(--vp-layout-max-width);
  margin: 0 auto;
  padding: 8px 40px;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 40px;
}

.content {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: center;
  font-size: 14px;
  font-weight: 500;
}

.badge {
  background: white;
  color: var(--vp-c-brand-1);
  padding: 1px 8px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: bold;
  text-transform: uppercase;
}

.link {
  color: white;
  text-decoration: underline;
  text-underline-offset: 4px;
  opacity: 0.9;
  transition: opacity 0.2s;
}

.link:hover {
  opacity: 1;
}

.close {
  position: absolute;
  right: 16px;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 4px;
  transition: background-color 0.2s;
  color: white;
}

.close:hover {
  background-color: rgba(255, 255, 255, 0.2);
}

.icon {
  width: 16px;
  height: 16px;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-100%);
}

@media (max-width: 640px) {
  .container {
    padding: 10px 40px 10px 16px;
  }
  .text {
    font-size: 13px;
  }
  .badge {
    display: none;
  }
}
</style>
