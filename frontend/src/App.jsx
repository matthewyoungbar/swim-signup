import React, { useState, useCallback } from 'react'
import PracticeCard from './components/PracticeCard'
import { usePractices } from './hooks/usePractices'
import { api } from './lib/api'

const STORAGE_KEY = 'swim_signup_user'

function loadUser() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY)) || null
  } catch {
    return null
  }
}

function saveUser(user) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(user))
}

export default function App() {
  const [user, setUser] = useState(loadUser)
  const [loginForm, setLoginForm] = useState({ name: '', email: '' })
  const [loginError, setLoginError] = useState(null)
  const [syncing, setSyncing] = useState(false)
  const [syncMsg, setSyncMsg] = useState(null)
  const [tab, setTab] = useState('upcoming') // 'upcoming' | 'mine'

  const { practices, loading, error, reload } = usePractices(user?.email)

  function handleLogin(e) {
    e.preventDefault()
    const name = loginForm.name.trim()
    const email = loginForm.email.trim().toLowerCase()
    if (!name || !email || !email.includes('@')) {
      setLoginError('Please enter a valid name and email.')
      return
    }
    const u = { name, email }
    setUser(u)
    saveUser(u)
  }

  function handleLogout() {
    setUser(null)
    localStorage.removeItem(STORAGE_KEY)
  }

  const handleSignup = useCallback(async (practiceId) => {
    await api.createSignup(practiceId, user.email, user.name)
    await reload()
  }, [user, reload])

  const handleCancel = useCallback(async (practiceId) => {
    await api.deleteSignup(practiceId, user.email)
    await reload()
  }, [user, reload])

  async function handleSync() {
    setSyncing(true)
    setSyncMsg(null)
    try {
      const result = await api.syncPractices()
      setSyncMsg(`Synced ${result.synced} practices from calendar.`)
      await reload()
    } catch (e) {
      setSyncMsg(`Sync failed: ${e.message}`)
    } finally {
      setSyncing(false)
    }
  }

  const myPractices = practices.filter(p => p.isSignedUp)
  const displayedPractices = tab === 'mine' ? myPractices : practices

  // ─── Not logged in ────────────────────────────────────────────────────────

  if (!user) {
    return (
      <div className="app">
        <header className="app-header">
          <div className="header-logo">
            <WaveIcon />
            <span>SwimSignup</span>
          </div>
        </header>
        <main className="login-page">
          <div className="login-card">
            <div className="login-icon"><WaveIcon size={40} /></div>
            <h1>Welcome back</h1>
            <p className="login-subtitle">Sign in to view and register for swim practices</p>
            <form onSubmit={handleLogin} className="login-form">
              <div className="field">
                <label htmlFor="name">Your name</label>
                <input
                  id="name"
                  type="text"
                  placeholder="Alex Smith"
                  value={loginForm.name}
                  onChange={e => setLoginForm(f => ({ ...f, name: e.target.value }))}
                  autoFocus
                />
              </div>
              <div className="field">
                <label htmlFor="email">Email address</label>
                <input
                  id="email"
                  type="email"
                  placeholder="alex@example.com"
                  value={loginForm.email}
                  onChange={e => setLoginForm(f => ({ ...f, email: e.target.value }))}
                />
              </div>
              {loginError && <p className="login-error">{loginError}</p>}
              <button type="submit" className="login-btn">Enter →</button>
            </form>
          </div>
        </main>
      </div>
    )
  }

  // ─── Logged in ────────────────────────────────────────────────────────────

  return (
    <div className="app">
      <header className="app-header">
        <div className="header-logo">
          <WaveIcon />
          <span>SwimSignup</span>
        </div>
        <div className="header-right">
          <button className="sync-btn" onClick={handleSync} disabled={syncing}>
            {syncing ? '⟳ Syncing…' : '⟳ Sync Calendar'}
          </button>
          <div className="user-pill" onClick={handleLogout} title="Click to sign out">
            <span className="user-avatar">{user.name[0].toUpperCase()}</span>
            <span className="user-name">{user.name}</span>
          </div>
        </div>
      </header>

      {syncMsg && (
        <div className={`sync-banner ${syncMsg.startsWith('Sync failed') ? 'error' : 'success'}`}>
          {syncMsg}
          <button onClick={() => setSyncMsg(null)}>✕</button>
        </div>
      )}

      <main className="main-content">
        <div className="tabs">
          <button
            className={`tab ${tab === 'upcoming' ? 'active' : ''}`}
            onClick={() => setTab('upcoming')}
          >
            All Practices
            <span className="tab-count">{practices.length}</span>
          </button>
          <button
            className={`tab ${tab === 'mine' ? 'active' : ''}`}
            onClick={() => setTab('mine')}
          >
            My Signups
            <span className="tab-count">{myPractices.length}</span>
          </button>
        </div>

        {loading && (
          <div className="loading-state">
            <div className="loading-wave">
              {[0,1,2,3,4].map(i => <span key={i} style={{ animationDelay: `${i * 0.1}s` }} />)}
            </div>
            <p>Loading practices…</p>
          </div>
        )}

        {error && !loading && (
          <div className="error-state">
            <p>⚠ {error}</p>
            <button onClick={reload}>Retry</button>
          </div>
        )}

        {!loading && !error && displayedPractices.length === 0 && (
          <div className="empty-state">
            {tab === 'mine'
              ? <p>You haven't signed up for any practices yet.</p>
              : <p>No upcoming practices. Try syncing the calendar.</p>
            }
          </div>
        )}

        <div className="practice-list">
          {displayedPractices.map(p => (
            <PracticeCard
              key={p.id}
              practice={p}
              onSignup={handleSignup}
              onCancel={handleCancel}
            />
          ))}
        </div>
      </main>
    </div>
  )
}

function WaveIcon({ size = 24 }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
      <path d="M2 12c1.5-2 3-2 4.5 0s3 2 4.5 0 3-2 4.5 0 3 2 4.5 0"/>
      <path d="M2 17c1.5-2 3-2 4.5 0s3 2 4.5 0 3-2 4.5 0 3 2 4.5 0"/>
      <path d="M2 7c1.5-2 3-2 4.5 0s3 2 4.5 0 3-2 4.5 0 3 2 4.5 0"/>
    </svg>
  )
}
