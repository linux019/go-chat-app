import React from 'react';
import PropTypes from 'prop-types';
import {serverAddress} from './App';
import ChatDialogue from './ChatDialogue';

import {Sidebar} from './items';

export const DataContext = React.createContext({});

class ChatApp extends React.Component {
    state = {
        activeChannelId: null,
        connected: false,
        channels: {},
        users: {},
        unreadChannels: {},
    };

    static propTypes = {
        userName: PropTypes.string.isRequired
    };

    constructor(props) {
        super(props);
        this.dialogueCallback = null;
    }

    componentDidMount() {
        this.openServerConnection();
    }

    openServerConnection = () => {
        const socket = new WebSocket(serverAddress);
        socket.onopen = this.onConnectionOpen;

        socket.onclose = event => {
            let reconnect = false;
            if (event.wasClean) {
                console.log('WS:DISCONNECTED');
            } else {
                reconnect = true;
                console.log('WS:DISCONNECTED (abort)');
            }
            console.log(`WS:DISCONNECTED (${event.code} ${event.reason})`);
            this.onConnectionClose(reconnect);
        };

        socket.onmessage = event => {
            console.log('DATA', event.data);
            try {
                const data = JSON.parse(event.data);
                if (data) {
                    this.onServerData(data);
                }
            } catch (e) {

            }
        };

        socket.onerror = function (error) {
            console.log('WS:ERROR', error);
        };
        this.socket = socket;
    };

    onConnectionOpen = () => {
        console.log('WS:OK');
        this.setState({connected: true});
        this.timeoutID = null;


        this.setName();
    };

    onConnectionClose = reconnect => {
        this.setState({connected: false});
        this.socket = null;
        if (reconnect) {
            this.timeoutID = setTimeout(this.openServerConnection, 10000);
        }
    };

    getChannels = () => this.sendCommand('GET_CHANNELS', null);
    setName = () => this.sendCommand('SET_USERNAME', this.props.userName);
    createChannel = (channel, isPublic) => this.sendCommand('CREATE_CHANNEL', {channel, isPublic});

    sendCommand = (command, data) => {
        if (this.socket) {
            this.socket.send(JSON.stringify({data, command}));
            console.log('WRITE: ', {data, command});
        }
    };

    onServerData = data => {
        const newState = {};
        const {activeChannel, unreadChannels} = this.state;

        const {channels, messages, message, channelName, users} = data;

        if (!this.socket) {
            return;
        }

        if (channels) {
            newState.channels = channels;
        }

        if (this.dialogueCallback) {
            (messages || message) && this.dialogueCallback(data);
        }

        if (message && activeChannel && channelName !== activeChannel.id) {
            newState.unreadChannels = {
                ...unreadChannels,
                ...{[channelName]: true}
            }
        }

        if (users) {
            newState.users = users;
        }

        if (Object.keys(newState).length) {
            this.setState(newState);
        }
    };

    componentWillUnmount() {
        this.socket.close();
        this.socket = null;
        if (this.timeoutID) {
            clearTimeout(this.timeoutID);
        }
    }

    setActiveChannel = id => {
        const unreadChannels = {...this.state.unreadChannels};
        delete unreadChannels[id];

        this.setState({
            activeChannel: {
                id,
                isP2P: false,
                peers: [],
            },
            unreadChannels
        });
    };

    setDialogueCallback = callback => {
        this.dialogueCallback = callback;
    };

    loadMessages = () => this.sendCommand('GET_CHANNEL_MESSAGES', {
        channel: this.state.activeChannel.id,
    });
    getUsersList = () => this.sendCommand('LIST_USERS', null);

    sendUserMessage = message => this.sendCommand('POST_MESSAGE', {
        ...{
            channel: this.state.activeChannel.id,
        },
        message
    });

    askForChannelName = e => {
        e.preventDefault();
        e.stopPropagation();
        const channel = window.prompt('Type a channel name');
        if (channel && channel.trim().length) {
            this.createChannel(channel, true);
        }
    };

    render() {
        const {userName} = this.props;
        const {channels, connected, activeChannel, unreadChannels, users} = this.state;
        const contextData = {
            userName, connected, channels, activeChannel, unreadChannels, users,
            askForChannelName: this.askForChannelName,
            setActiveChannel: this.setActiveChannel,
            getUsersList: this.getUsersList,
        };

        return (
            <DataContext.Provider
                value={contextData}>
                <Sidebar/>
                {
                    activeChannelId && channels[activeChannelId] &&
                    <ChatDialogue key={activeChannelId}
                                  activeChannelId={activeChannelId}
                                  setCallback={this.setDialogueCallback}
                                  sendUserMessage={this.sendUserMessage}
                                  loadMessages={this.loadMessages}
                    />
                }
            </DataContext.Provider>
        )
    }
}

export default ChatApp