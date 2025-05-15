import React, { useEffect, useState, useRef } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

function Modal({ show, onClose, event }) {
  if (!show) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center z-50">
      <div className="bg-white rounded p-6 w-96 max-w-full">
        <h2 className="text-xl font-semibold mb-4">Attack Details</h2>
        <pre className="whitespace-pre-wrap break-words">{JSON.stringify(event, null, 2)}</pre>
        <button
          className="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          onClick={onClose}
        >
          Close
        </button>
      </div>
    </div>
  );
}

function App() {
  const [events, setEvents] = useState([]);
  const [filteredEvents, setFilteredEvents] = useState([]);
  const [eventCounts, setEventCounts] = useState([]);
  const [systemActive, setSystemActive] = useState(false);
  const [loading, setLoading] = useState(false);
  const [filter, setFilter] = useState('all');
  const [modalEvent, setModalEvent] = useState(null);
  const [modalVisible, setModalVisible] = useState(false);
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);

  // Fetch initial system status from Agent
  useEffect(() => {
    fetch('http://localhost:8081/status')
      .then(res => res.json())
      .then(data => {
        if (data.status && data.status.toLowerCase().includes('active')) {
          setSystemActive(true);
        } else {
          setSystemActive(false);
        }
      })
      .catch(err => {
        console.error('Error fetching agent status:', err);
      });
  }, []);

  // Fetch initial events from backend Engine
  useEffect(() => {
    fetch('http://localhost:8000/api/events')
      .then(res => res.json())
      .then(data => {
        setEvents(data);
        setFilteredEvents(data);
        updateEventCounts(data);
      })
      .catch(err => console.error('Error fetching initial events:', err));
  }, []);

  // Filter events based on filter state
  useEffect(() => {
    if (filter === 'all') {
      setFilteredEvents(events);
    } else {
      setFilteredEvents(events.filter(e => e.type.toLowerCase() === filter));
    }
  }, [filter, events]);

  // WebSocket connection with auto reconnect
  useEffect(() => {
    function connect() {
      console.log('Connecting to WebSocket...');
      const ws = new WebSocket('ws://localhost:8000/ws');
      wsRef.current = ws;

      ws.onopen = () => {
        console.log('Connected to Engine WebSocket');
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          console.log('Received WebSocket data:', data);
          setEvents(prev => {
            const updated = [...prev, data];
            updateEventCounts(updated);
            return updated;
          });
        } catch (e) {
          console.error('Error parsing WebSocket message:', e);
        }
      };

      ws.onclose = () => {
        console.log('WebSocket connection closed, attempting to reconnect in 3 seconds...');
        reconnectTimeoutRef.current = setTimeout(() => {
          connect();
        }, 3000);
      };

      ws.onerror = (err) => {
        console.error('WebSocket error:', err);
        ws.close();
      };
    }

    connect();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, []);

  const updateEventCounts = (events) => {
    const countsMap = {};
    events.forEach(event => {
      const timestamp = new Date(event.timestamp);
      const timeKey = timestamp.toISOString().slice(0, 16); // up to minute precision
      countsMap[timeKey] = (countsMap[timeKey] || 0) + 1;
    });
    const countsArray = Object.entries(countsMap).map(([time, count]) => ({ time, count }));
    countsArray.sort((a, b) => new Date(a.time) - new Date(b.time));
    setEventCounts(countsArray);
  };

  const toggleSystem = async () => {
    setLoading(true);
    try {
      const newState = !systemActive;
      // Call agent endpoint
      const agentResponse = await fetch(`http://localhost:8081/${newState ? 'activate' : 'deactivate'}`, { method: 'POST' });
      if (!agentResponse.ok) {
        throw new Error('Agent request failed');
      }
      // Call engine endpoint
      const engineResponse = await fetch(`http://localhost:8000/api/engine/${newState ? 'activate' : 'deactivate'}`, { method: 'POST' });
      if (!engineResponse.ok) {
        throw new Error('Engine request failed');
      }
      setSystemActive(newState);
    } catch (error) {
      console.error('Error toggling system:', error);
      alert('Failed to toggle system status. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const openModal = (event) => {
    setModalEvent(event);
    setModalVisible(true);
  };

  const closeModal = () => {
    setModalVisible(false);
    setModalEvent(null);
  };

  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4">SecureShield Dashboard</h1>
      <button
        id="toggleSystemBtn"
        type="button"
        className={`mb-4 text-white text-sm font-semibold px-4 py-1 rounded ${systemActive ? 'bg-green-600' : 'bg-red-600'}`}
        aria-pressed={systemActive}
        aria-label="Toggle system status"
        onClick={toggleSystem}
        disabled={loading}
      >
        {systemActive ? 'System On' : 'System Off'}
      </button>

      <div className="mb-4">
        <label htmlFor="filter" className="mr-2 font-semibold">Filter Attack Type:</label>
        <select
          id="filter"
          value={filter}
          onChange={e => setFilter(e.target.value)}
          className="border border-gray-300 rounded px-2 py-1"
        >
          <option value="all">All</option>
          <option value="sql-injection">SQL Injection</option>
          <option value="web-shell">Web Shell</option>
          <option value="ddos">DDoS</option>
        </select>
      </div>

      <div className="overflow-x-auto mb-6">
        <table className="min-w-full divide-y divide-gray-200 text-sm">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-2 text-left font-semibold text-gray-700">Time</th>
              <th className="px-4 py-2 text-left font-semibold text-gray-700">Type</th>
              <th className="px-4 py-2 text-left font-semibold text-gray-700">Source IP</th>
              <th className="px-4 py-2 text-left font-semibold text-gray-700">Content</th>
              <th className="px-4 py-2 text-left font-semibold text-gray-700">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredEvents.length === 0 ? (
              <tr>
                <td colSpan="5" className="text-center py-4 text-gray-500">Belum ada serangan terdeteksi</td>
              </tr>
            ) : (
              filteredEvents.map((event, idx) => (
                <tr key={idx} className="hover:bg-gray-100">
                  <td className="px-4 py-2">{new Date(event.timestamp).toLocaleString()}</td>
                  <td className="px-4 py-2">{event.type}</td>
                  <td className="px-4 py-2">{event.source}</td>
                  <td className="px-4 py-2">{event.content}</td>
                  <td className="px-4 py-2">
                    <button
                      className="text-blue-600 hover:underline"
                      onClick={() => openModal(event)}
                    >
                      View
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <div style={{ width: '100%', height: 300 }}>
        <h2 className="text-lg font-semibold mb-2">Attack Trend (Events per Minute)</h2>
        <ResponsiveContainer>
          <LineChart data={eventCounts} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="time" />
            <YAxis allowDecimals={false} />
            <Tooltip />
            <Legend />
            <Line type="monotone" dataKey="count" stroke="#8884d8" activeDot={{ r: 8 }} />
          </LineChart>
        </ResponsiveContainer>
      </div>

      <Modal show={modalVisible} onClose={closeModal} event={modalEvent} />
    </div>
  );
}

export default App;
