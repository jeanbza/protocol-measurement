import React from 'react'

class RunResults extends React.Component {
    constructor(props) {
        super(props)

        const {selectedRun} = props

        const runInterval = setInterval(() => fetch(new Request(`/runs/${selectedRun}`))
            .then(resp => resp.json())
            .then(run => this.setState({run}))
            .catch(err => {
                console.error(err)
                clearInterval(runInterval)
            }), 200)

        const progressInterval = setInterval(() => {
            console.log('looking up', selectedRun)
            fetch(new Request(`/runs/${selectedRun}/results`))
                .then(resp => resp.json())
                .then(progress => this.updateProgress.bind(this)(selectedRun, progress))
                .catch(err => {
                    console.error(err)
                    clearInterval(progressInterval)
                }), 200
        })

        this.state = {
            runInterval,
            progressInterval,
            run: {},
            progress: {},
        }
    }

    componentWillReceiveProps(nextProps) {
        const {selectedRun} = nextProps

        clearInterval(this.state.runInterval)
        clearInterval(this.state.progressInterval)

        const runInterval = setInterval(() => fetch(new Request(`/runs/${selectedRun}`))
            .then(resp => resp.json())
            .then(run => this.setState({run}))
            .catch(err => console.error(err)), 200)

        const progressInterval = setInterval(() => {
            console.log('looking up', selectedRun)
            fetch(new Request(`/runs/${selectedRun}/results`))
                .then(resp => resp.json())
                .then(progress => this.updateProgress.bind(this)(selectedRun, progress))
                .catch(err => console.error(err)), 200
        })

        this.setState({
            runInterval,
            progressInterval,
            run: {},
            progress: {},
        })
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

        const progressBars = Object.keys(progress)
            .filter(k => progress[k])
            .map(k => <div key={k}>
                <label>{k}
                    <small>avg {Math.round(progress[k]['avgTravelTime'])}ms</small>
                </label>
                <progress value={progress[k]['count']} max={totalMessages}/>
            </div>)

        return <div className="results">
            <div>Sent messages: {totalMessages}</div>
            {progressBars}
        </div>
    }

    updateProgress(selectedRun, newProgress) {
        if (selectedRun !== this.props.selectedRun) {
            return
        }

        this.setState({
            progress: {
                'http': false,
                'udp': false,
                'quic': false,
                'websocket': false,
                'grpc-streaming': false,
                'grpc-unary': false,
                ...this.state.progress,
                ...newProgress,
            }
        })
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