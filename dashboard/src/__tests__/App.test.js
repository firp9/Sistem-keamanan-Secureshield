import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import App from '../App';

describe('SecureShield Dashboard Toggle', () => {
  beforeEach(() => {
    // Mock fetch for status and toggle POST requests
    global.fetch = jest.fn((url, options) => {
      if (url === 'http://secureshield_agent:8080/status') {
        return Promise.resolve({
          ok: true,
          json: () => Promise.resolve({ status: 'inactive' }),
        });
      }
      if (url === 'http://secureshield_agent:8080/activate' && options.method === 'POST') {
        return Promise.resolve({ ok: true });
      }
      if (url === 'http://secureshield_agent:8080/deactivate' && options.method === 'POST') {
        return Promise.resolve({ ok: true });
      }
      if (url.startsWith('http://engine:8000/api/engine/')) {
        return Promise.resolve({ ok: true });
      }
      return Promise.reject(new Error('Unknown URL'));
    });
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  test('toggle button sends POST requests to agent and updates UI', async () => {
    render(<App />);

    // Wait for initial status fetch
    await waitFor(() => expect(global.fetch).toHaveBeenCalledWith('http://secureshield_agent:8080/status'));

    const toggleButton = screen.getByRole('button', { name: /toggle system status/i });

    // Initial state should be off
    expect(toggleButton).toHaveTextContent(/system off/i);

    // Click to activate
    fireEvent.click(toggleButton);

    // Wait for POST requests to be called
    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith('http://secureshield_agent:8080/activate', { method: 'POST' });
      expect(global.fetch).toHaveBeenCalledWith('http://engine:8000/api/engine/activate', { method: 'POST' });
    });

    // Button text should update to "System On"
    expect(toggleButton).toHaveTextContent(/system on/i);

    // Click to deactivate
    fireEvent.click(toggleButton);

    // Wait for POST requests to be called
    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith('http://secureshield_agent:8080/deactivate', { method: 'POST' });
      expect(global.fetch).toHaveBeenCalledWith('http://engine:8000/api/engine/deactivate', { method: 'POST' });
    });

    // Button text should update to "System Off"
    expect(toggleButton).toHaveTextContent(/system off/i);
  });
});
