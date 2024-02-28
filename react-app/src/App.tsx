import React, { useState, useEffect } from 'react';

interface AppProps {}

interface CacheData {
  key: string;
  value: string;
}

const App: React.FC<AppProps> = () => {
  const [key, setKey] = useState('');
  const [value, setValue] = useState('');
  const [data, setData] = useState<CacheData | null>(null);

  const getFromCache = async () => {
    try {
      const response = await fetch(`http://localhost:8086/get?key=${key}`, {
        headers: { 'Content-Type': 'application/json' },
      });
      const data = await response.json();
      setData(data);
    } catch (error) {
      console.error('Error getting from cache:', error);
      // Handle errors, e.g., display an error message to the user
    }
  };

  const getAllFromCache = async () => {
    try {
      const response = await fetch('http://localhost:8086/get-all', {
        headers: { 'Content-Type': 'application/json' },
      });
      const data = await response.json();
      setData(data);
    } catch (error) {
      console.error('Error getting all from cache:', error);
      // Handle errors
    }
  };

  const setToCache = async () => {
    try {
      const response = await fetch('http://localhost:8086/set', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key, value }),
      });

      if (!response.ok) {
        throw new Error('Failed to set data to cache');
      }

      console.log('Data set successfully!');
      // Optionally display a success message
    } catch (error) {
      console.error('Error setting data:', error);
      // Handle errors
    }
  };

  return (
    <div className="App">
      <h1>LRU Cache Manager</h1>
      <div className="input-container">
        <input
          type="text"
          placeholder="Key"
          value={key}
          onChange={(e) => setKey(e.target.value)}
        />
        <button onClick={getFromCache} disabled={!key}>
          Get
        </button>
        {key && (
          <button onClick={getAllFromCache} disabled={!key}>
            Get-All
          </button>
        )}
      </div>
      <div className="input-container">
        <input
          type="text"
          placeholder="Key"
          value={key}
          onChange={(e) => setKey(e.target.value)}
        />
        <input
          type="text"
          placeholder="Value"
          value={value}
          onChange={(e) => setValue(e.target.value)}
        />
        <button onClick={setToCache} disabled={!key || !value}>
          Set
        </button>
      </div>
      {data && (
        <div className="result-box">

          <b>Key:</b> {data.key} <br />
          <b>Value:</b> {data.value}
        </div>
      )}
    </div>
  );
};
export default App;
