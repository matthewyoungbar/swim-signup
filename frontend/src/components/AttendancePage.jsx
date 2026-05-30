import React, { useEffect, useState, useCallback } from 'react'
import './AttendancePage.css'
import { api } from '../lib/api'

function displayName(u) {
  return u.preferredName
    ? `${u.preferredName} ${u.lastName}`
    : `${u.firstName} ${u.lastName}`
}

function formatDate(iso) {
  return new Date(iso).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })
}

function formatTime(iso) {
  return new Date(iso).toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' })
}

export default function AttendancePage({ user }) {
  const [practices, setPractices]       = useState([])
  const [swimmers, setSwimmers]         = useState([])
  const [selected, setSelected]         = useState(null)
  const [attendeeEmails, setAttendeeEmails] = useState(new Set())
  const [notes, setNotes]               = useState('')
  const [loading, setLoading]           = useState(true)
  const [loadingAtt, setLoadingAtt]     = useState(false)
  const [saving, setSaving]             = useState(false)
  const [saveStatus, setSaveStatus]     = useState(null) // 'saved' | error string
  const [search, setSearch]             = useState('')

  useEffect(() => {
    Promise.all([api.getAllPractices(), api.listSwimmers()])
      .then(([ps, sw]) => {
        const myPractices = user.isAdmin
          ? ps
          : ps.filter(p => p.coachId === 'USER#' + user.email)
        myPractices.sort((a, b) => new Date(b.startTime) - new Date(a.startTime))
        setPractices(myPractices)
        setSwimmers(sw)
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [user])

  const selectPractice = useCallback(async (practice) => {
    setSelected(practice)
    setSaveStatus(null)
    setSearch('')
    setLoadingAtt(true)
    try {
      const att = await api.getAttendance(practice.id)
      setAttendeeEmails(new Set((att.attendees || []).map(a => a.email)))
      setNotes(att.notes || '')
    } catch {
      setAttendeeEmails(new Set())
      setNotes('')
    } finally {
      setLoadingAtt(false)
    }
  }, [])

  function toggleAttendee(email) {
    setAttendeeEmails(prev => {
      const next = new Set(prev)
      next.has(email) ? next.delete(email) : next.add(email)
      return next
    })
  }

  async function handleSave() {
    setSaving(true)
    setSaveStatus(null)
    try {
      const attendees = swimmers
        .filter(s => attendeeEmails.has(s.email))
        .map(s => ({ email: s.email, name: displayName(s) }))
      await api.saveAttendance(selected.id, { attendees, notes })
      setSaveStatus('saved')
    } catch (e) {
      setSaveStatus(e.message)
    } finally {
      setSaving(false)
    }
  }

  // ── Attendance form ────────────────────────────────────────────────────────
  if (selected) {
    const q = search.trim().toLowerCase()
    const filtered = q
      ? swimmers.filter(s => displayName(s).toLowerCase().includes(q) || s.email.toLowerCase().includes(q))
      : swimmers

    return (
      <div className="attendance-page">
        <button className="account-back" onClick={() => setSelected(null)}>← Practices</button>

        <div className="att-practice-header">
          <h2 className="att-practice-title">{selected.title}</h2>
          <p className="att-practice-meta">
            {formatDate(selected.startTime)} · {formatTime(selected.startTime)}–{formatTime(selected.endTime)}
            {selected.location ? ` · ${selected.location}` : ''}
          </p>
        </div>

        {loadingAtt ? (
          <p className="account-muted">Loading attendance…</p>
        ) : (
          <>
            <div className="att-section">
              <div className="att-section-header">
                <span className="att-section-title">Attendance</span>
                <span className="att-count">{attendeeEmails.size} present</span>
                <input
                  className="att-search"
                  type="search"
                  placeholder="Filter swimmers…"
                  value={search}
                  onChange={e => setSearch(e.target.value)}
                />
              </div>
              <div className="att-swimmer-list">
                {filtered.map(s => (
                  <label key={s.email} className="att-swimmer-row">
                    <input
                      type="checkbox"
                      className="att-checkbox"
                      checked={attendeeEmails.has(s.email)}
                      onChange={() => toggleAttendee(s.email)}
                    />
                    <span className="att-swimmer-name">{displayName(s)}</span>
                    <span className="att-swimmer-email">{s.email}</span>
                  </label>
                ))}
                {filtered.length === 0 && <p className="account-muted">No swimmers match.</p>}
              </div>
            </div>

            <div className="att-section">
              <div className="att-section-title" style={{ marginBottom: 8 }}>Notes</div>
              <textarea
                className="att-notes"
                placeholder="Practice notes, observations…"
                value={notes}
                onChange={e => setNotes(e.target.value)}
                rows={4}
              />
            </div>

            <div className="att-save-row">
              <button className="login-btn" style={{ width: 'auto', padding: '10px 28px' }} onClick={handleSave} disabled={saving}>
                {saving ? 'Saving…' : 'Save attendance'}
              </button>
              {saveStatus === 'saved' && <span className="att-save-ok">Saved</span>}
              {saveStatus && saveStatus !== 'saved' && <span className="att-save-err">{saveStatus}</span>}
            </div>
          </>
        )}
      </div>
    )
  }

  // ── Practice list ──────────────────────────────────────────────────────────
  return (
    <div className="attendance-page">
      <h2 className="admin-title" style={{ marginBottom: 20 }}>
        {user.isAdmin ? 'Attendance' : 'My Practices'}
      </h2>

      {loading ? (
        <div className="loading-state">
          <div className="loading-wave">
            {[0,1,2,3,4].map(i => <span key={i} style={{ animationDelay: `${i * 0.1}s` }} />)}
          </div>
          <p>Loading practices…</p>
        </div>
      ) : practices.length === 0 ? (
        <div className="empty-state">
          <p>{user.isAdmin ? 'No practices found.' : 'No practices assigned to you yet.'}</p>
        </div>
      ) : (
        <div className="att-practice-list">
          {practices.map(p => (
            <button key={p.id} className="att-practice-row" onClick={() => selectPractice(p)}>
              <div className="practice-date-badge" style={{ minWidth: 52 }}>
                <span className="day-name">{new Date(p.startTime).toLocaleDateString('en-US', { weekday: 'short' })}</span>
                <span className="day-num">{new Date(p.startTime).getDate()}</span>
                <span className="month">{new Date(p.startTime).toLocaleDateString('en-US', { month: 'short' })}</span>
              </div>
              <div className="att-practice-info">
                <span className="att-practice-name">{p.title}</span>
                <span className="att-practice-time">
                  {formatTime(p.startTime)}–{formatTime(p.endTime)}
                  {p.location ? ` · ${p.location}` : ''}
                  {p.coachName ? ` · ${p.coachName}` : ''}
                </span>
              </div>
              <span className="att-chevron">›</span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
