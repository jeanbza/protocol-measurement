import React from 'react'

class RunResults extends React.PureComponent {
    constructor(props) {
        super(props)

        const {selectedRun} = props

        const runInterval = setInterval(() => fetch(new Request(`/runs/${selectedRun}`))
            .then(resp => resp.json())
            .then(run => this.setState({run}))
            .catch(err => console.error(err)), 200)

        const progressInterval = setInterval(() => fetch(new Request(`/runs/${selectedRun}/results`))
            .then(resp => resp.json())
            .then(progress => this.setState({progress}))
            .catch(err => console.error(err)), 200)

        this.state = {
            runInterval,
            progressInterval,
            run: {},
            progress: {},
        }
    }

    componentWillUnmount() {
        clearInterval(this.state.runInterval)
        clearInterval(this.state.progressInterval)
    }

    componentWillUpdate(nextProps, nextState) {
        if (nextState.run.finishedCreating) {
            clearInterval(this.state.runInterval)
        }
    }

    render() {
        const {run: {id, totalMessages}, progress} = this.state

        const fullProgress = {
            'http': false,
            'udp': false,
            'quic': false,
            'websocket': false,
            'grpc-streaming': false,
            'grpc-unary': false,
            ...progress
        }

        const progressBars = Object.keys(fullProgress).map(k => <div key={k}>
            <label>{k}</label>
            <progress value={fullProgress[k]} max={totalMessages}/>
        </div>)

        return <div className="results-area">
            <div>Sent messages: {totalMessages}</div>
            {progressBars}
        </div>
    }
}

export default class ResultsArea extends React.Component {
    render() {
        const {selectedRun} = this.props

        let content = <div/>
        let subtitle = <small>No run selected</small>

        if (selectedRun) {
            content = <RunResults selectedRun={selectedRun}/>
            subtitle = <small>Viewing run {selectedRun}</small>
        }

        return <div className="results-area">
            <h3>Results area</h3>
            {subtitle}
            {content}
        </div>
    }
}