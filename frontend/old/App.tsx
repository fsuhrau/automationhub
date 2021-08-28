import React, { useCallback, useEffect, useState } from 'react';
import logo from './logo.svg';
import './App.css';

const socket = new WebSocket("ws://10.35.111.51:8080/ws");

function App() {
  const [message, setMessage] = useState('')
  const [inputValue, setInputValue] = useState('')

  useEffect(() =>{
    socket.onopen = () => {
      setMessage('Connected')
    };

    socket.onmessage = (e) => {
      setMessage("Get message from ; " + e.data)
    };

    return () => {
      socket.close()
    };
  }, [])

  const handleClick = useCallback((e) => {
    e.preventDefault()
    socket.send(JSON.stringify({
      message: e.target.value
    }))
  }, [inputValue])

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue(e.target.value)
  }, []);

  return (
    <div className="App">
      <input id="input" type="text" value={inputValue} onChange={handleChange} />
      <button onClick={handleClick}>Send</button>
      <pre>{message}</pre>
    </div>
  );
}

export default App;
