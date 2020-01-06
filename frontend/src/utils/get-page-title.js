import defaultSettings from '@/settings'

const title = defaultSettings.title || 'Raft Demo'

export default function getPageTitle(pageTitle) {
  if (pageTitle) {
    return `${pageTitle} - ${title}`
  }
  return `${title}`
}
