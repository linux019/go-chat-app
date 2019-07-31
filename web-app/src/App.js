import React, {useState} from 'react';
import './App.scss';
import ChatApp from './ChatApp';
import {WelcomeScreen} from './items';

export const serverAddress = process.env.NODE_ENV === 'production' ? `wss://${window.location.hostname}/ws` : 'ws://localhost:4488/ws';

function App() {
    const [{name, openChat}, setData] = useState({name: '', openChat: false});
    return (
        openChat
            ? <ChatApp userName={name}/>
            : <WelcomeScreen {...{name, setData}}/>
    );
}

export default App;
