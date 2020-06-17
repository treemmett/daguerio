import {
  BeforeInsert,
  BeforeUpdate,
  Column,
  CreateDateColumn,
  Entity,
  PrimaryColumn,
} from 'typeorm';
import {
  IsNotEmpty,
  IsUUID,
  Max,
  Min,
  validateOrReject,
} from 'class-validator';

@Entity({ name: 'photos' })
export default class Photo {
  @PrimaryColumn('uuid')
  @IsUUID('4')
  public readonly id: string;

  @Column({ type: 'bigint' })
  @IsNotEmpty()
  @Min(1)
  @Max(10000000000) // 10 gb
  public size: number;

  @CreateDateColumn()
  public readonly uploadedTime: Date;

  public constructor(id: string) {
    this.id = id;
  }

  @BeforeInsert()
  @BeforeUpdate()
  public async validate(): Promise<void> {
    await validateOrReject(this);
  }
}
