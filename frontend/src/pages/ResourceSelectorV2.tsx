/*
 * Copyright 2023 The Kubeflow Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Separate the resource selector between v1 and v2 to avoid breaking current v1 behavior
// TODO(jlyaoyuli): consider to merge 2 selectors together (change updatedSelection() in v1)

import * as React from 'react';
import CustomTable, { Column, Row } from 'src/components/CustomTable';
import Toolbar, { ToolbarActionMap } from 'src/components/Toolbar';
import { ListRequest } from 'src/lib/Apis';
import { RouteComponentProps } from 'react-router-dom';
import { logger, errorToMessage, formatDateString } from 'src/lib/Utils';
import { DialogProps } from 'src/components/Router';

interface BaseResponse {
  resources: BaseResource[];
  nextPageToken: string;
}

export interface BaseResource {
  id?: string;
  created_at?: Date;
  description?: string;
  name?: string;
  error?: string;
  nameSpace?: string;
}

export interface ResourceSelectorV2Props extends RouteComponentProps {
  listApi: (...args: any[]) => Promise<BaseResponse>;
  columns: Column[];
  emptyMessage: string;
  filterLabel: string;
  initialSortColumn: any;
  selectionChanged: (selectedId: string) => void;
  title?: string;
  toolbarActionMap?: ToolbarActionMap;
  updateDialog: (dialogProps: DialogProps) => void;
}

interface ResourceSelectorV2State {
  resources: BaseResource[];
  rows: Row[];
  selectedIds: string[];
  toolbarActionMap: ToolbarActionMap;
}

class ResourceSelectorV2 extends React.Component<ResourceSelectorV2Props, ResourceSelectorV2State> {
  protected _isMounted = true;

  constructor(props: any) {
    super(props);

    this.state = {
      resources: [],
      rows: [],
      selectedIds: [],
      toolbarActionMap: (props && props.toolbarActionMap) || {},
    };
  }

  public render(): JSX.Element {
    const { rows, selectedIds, toolbarActionMap } = this.state;
    const { columns, title, filterLabel, emptyMessage, initialSortColumn } = this.props;

    return (
      <React.Fragment>
        {title && <Toolbar actions={toolbarActionMap} breadcrumbs={[]} pageTitle={title} />}

        <CustomTable
          columns={columns}
          rows={rows}
          selectedIds={selectedIds}
          useRadioButtons={true}
          updateSelection={this._selectionChanged.bind(this)}
          filterLabel={filterLabel}
          initialSortColumn={initialSortColumn}
          reload={this._load.bind(this)}
          emptyMessage={emptyMessage}
        />
      </React.Fragment>
    );
  }

  public componentWillUnmount(): void {
    this._isMounted = false;
  }

  protected setStateSafe(newState: Partial<ResourceSelectorV2State>, cb?: () => void): void {
    if (this._isMounted) {
      this.setState(newState as any, cb);
    }
  }

  protected _selectionChanged(selectedIds: string[]): void {
    if (!Array.isArray(selectedIds) || selectedIds.length !== 1) {
      logger.error(`${selectedIds.length} resources were selected somehow`, selectedIds);
      return;
    }
    this.props.selectionChanged(selectedIds[0]);
    this.setStateSafe({ selectedIds });
  }

  protected async _load(request: ListRequest): Promise<string> {
    let nextPageToken = '';
    try {
      const response = await this.props.listApi(
        request.pageToken,
        request.pageSize,
        request.sortBy,
        request.filter,
      );

      this.setStateSafe({
        resources: response.resources,
        rows: this._resourcesToRow(response.resources),
      });

      nextPageToken = response.nextPageToken;
    } catch (err) {
      const errorMessage = await errorToMessage(err);
      this.props.updateDialog({
        buttons: [{ text: 'Dismiss' }],
        content: 'List request failed with:\n' + errorMessage,
        title: 'Error retrieving resources',
      });
      logger.error('Could not get requested list of resources', errorMessage);
    }
    return nextPageToken;
  }

  protected _resourcesToRow(resources: BaseResource[]): Row[] {
    return resources.map(
      r =>
        ({
          error: (r as any).error,
          id: r.id!,
          otherFields: [r.name, r.description, formatDateString(r.created_at)],
        } as Row),
    );
  }
}

export default ResourceSelectorV2;