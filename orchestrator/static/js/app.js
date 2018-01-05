import '../css/app.scss' // required for wepback to build css

import React from 'react'
import {render} from 'react-dom'

import Sidebar from './components/Sidebar'
import ResultsArea from './components/ResultsArea'

class App extends React.Component {
    constructor(props) {
        super(props)

        this.state = {
            selectedRun: undefined,
        }
    }

    onRunChangeHandler(newRun) {
        this.setState({
            selectedRun: newRun,
        })
    }

    render() {
        const {selectedRun} = this.state

        return <div className="page">
            <Sidebar selectedRun={selectedRun} onRunChange={this.onRunChangeHandler.bind(this)}/>
            <ResultsArea selectedRun={selectedRun}/>
        </div>
    }
}

render(<App/>, document.getElementById('root'))