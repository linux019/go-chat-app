import React from 'react';
import './App.scss';
// Teddie Miller
// Reece Sharp
// Hamish Matthews
// Charles Burns
// Eric Palmer
// Caelan Green
// Wilfred Khan
// Alex Allen
// Charlie Stevens
// Harley Robertson
function App() {
    return (
        <>
            <div className='sidebar'>
                <div className="card blue-grey darken-1">
                    <div className="card-content white-text">
                        <span className="card-title">Card Title</span>
                        <p>I am a very simple card</p>
                    </div>
                    <div className="card-action">
                        <a href="#">
                            <i className={'material-icons tiny'}>check_circle</i>
                            This is a link</a>
                    </div>
                </div>

                <ul className="collection with-header">
                    <li className="collection-header"><h6>First Names</h6></li>
                    <li className="collection-item">Alvin</li>
                    <li className="collection-item">Alvin</li>
                    <li className="collection-item">Alvin</li>
                    <li className="collection-item">Alvin</li>
                </ul>
            </div>
            <div className="dialogue">
                content

            </div>
        </>
    );
}

export default App;
