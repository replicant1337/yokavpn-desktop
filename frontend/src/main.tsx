import { render } from 'preact'
import App from './App'
import { TranslationProvider } from './i18n'
import 'react-flagpack/dist/style.css'
import './style.css'

render(
  <TranslationProvider>
    <App />
  </TranslationProvider>, 
  document.getElementById('app')!
)