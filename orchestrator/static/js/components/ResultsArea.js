import React from 'react'

export default class ResultsArea extends React.Component {
    render() {
        const {selectedSet} = this.props

        let content = <small>No set selected!</small>

        if (selectedSet) {
            content = <small>Viewing set {selectedSet}</small>
        }

        return <div className="results-area">
            <h3>Results area</h3>
            {content}
            <div className="results">
            </div>
        </div>
    }
}