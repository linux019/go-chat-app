import moment from 'moment';
import React from 'react';
import PropTypes from 'prop-types';
import {DataContext} from './ChatApp';
import {ChannelPublicIcon} from './items';

class ChatDialogue extends React.Component {
    state = {
        messages: [],
        text: '',
    };

    static propTypes = {
        activeChannelId: PropTypes.string.isRequired,
        loadMessages: PropTypes.func.isRequired,
        setCallback: PropTypes.func.isRequired,
        sendUserMessage: PropTypes.func.isRequired,
    };

    componentDidMount() {
        this.props.setCallback(this.storeMessages);
        this.props.loadMessages();
    }

    storeMessages = ({messages, message, channelName}) => {
        if (message && channelName === this.props.activeChannelId) {
            messages = [...this.state.messages];
            messages.push(message);
        }
        if (messages) {
            this.setState({messages});
        }
    };

    componentWillUnmount() {
        this.props.setCallback(null);
    }

    onTextChange = e => this.setState({text: e.target.value});

    onSubmit = e => {
        if (e.key === 'Enter') {
            e.preventDefault();
            e.stopPropagation();
            this.props.sendUserMessage(this.state.text);
            this.setState({text: ''});
        }
    };

    render() {
        const {messages, text} = this.state;
        const {activeChannelId} = this.props;
        return (
            <div className="dialogue">
                <div className='chat'>
                    <DataContext.Consumer>
                        {
                            ({channels}) => {
                                const {isSelf, isPublic} = channels[activeChannelId];
                                return (
                                    <div
                                        className='chat-header'>
                                        {
                                            !isSelf && <ChannelPublicIcon isPublic={isPublic}/>
                                        }
                                        &nbsp;
                                        {isSelf ? 'Save Your Messages Here' : activeChannelId}
                                    </div>
                                )
                            }
                        }
                    </DataContext.Consumer>
                    <div className={'messages'}>
                        {
                            messages.length > 0
                                ? messages.map(message =>
                                    <ChatMessage key={`${message.sender}-${message.time}`} {...message}/>)
                                : <p className={'center'}>No messages yet</p>
                        }
                    </div>
                    <div className={'text-input'}>
                        <textarea value={text}
                                  onChange={this.onTextChange}
                                  onKeyDown={this.onSubmit}
                        />
                    </div>
                </div>
            </div>
        )
    }
}

function dateFormat(unixtime) {
    const msgDate = new Date();
    msgDate.setTime(unixtime * 1000);
    const isSame = msgDate.getFullYear() === new Date().getFullYear();
    const momentDate = moment.unix(unixtime);
    const format = {
        sameDay: 'h:mm A',
        lastDay: '[Yesterday], h:mm A',
        lastWeek: 'MMM D, h:mm A',
        sameElse: 'MMM D, h:mm A'
    };
    return isSame ? momentDate.calendar(null, format) : momentDate.format('MMM D, YYYY h:mm A');
}

const ChatMessage = ({time, sender, message}) => (
    <div className='message'>
        <div className='header'>
            <span className='sender'>{sender}</span>
            <span className='time'>{dateFormat(time)}</span>
        </div>
        <div className='text'>{message}</div>
    </div>
);

export default ChatDialogue